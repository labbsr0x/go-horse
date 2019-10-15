package tests

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/filters/list"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/web/config-web"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/spf13/viper"
	"gotest.tools/icmd"
)

var app *web.Server

func init() {

	os.Setenv("docker-api-version", "1.39")
	os.Setenv("docker-sock-url", "unix:///var/run/docker.sock")
	os.Setenv("target-host-name", "http://go-horse")
	os.Setenv("log-level", "true")
	os.Setenv("pretty-log", "true")
	os.Setenv("port", ":8080")
	os.Setenv("js-filter-path", "/app/go-horse/filters")
	os.Setenv("go-plugins-path", "/app/go-horse/plugins")

	go func() {
		list.Reload()
		webBuilder := new(web.WebBuilder).InitFromViper(viper.GetViper())
		app = new(web.Server).InitFromWebBuilder(webBuilder)
		app.Run()
	}()
	SetDefaultFailureMode(FailureContinues)
}

// https://github.com/docker/cli/blob/master/e2e/container/run_test.go
// https://github.com/smartystreets/goconvey/
// https://github.com/ory/dockertest

func TestRun(t *testing.T) {
	Convey("docker run --name e2e-test-container -d redis", t, func() {
		result := icmd.RunCommand("docker", "-H", "tcp://localhost:8080", "run", "--name", "e2e-test-container", "-d", "redis")
		So(len(strings.TrimSpace(result.Stdout())), ShouldEqual, 64)
		So(result.ExitCode, ShouldEqual, 0)
	})
}

func TestRemove(t *testing.T) {
	Convey("docker run -f e2e-test-container", t, func() {
		result := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", "e2e-test-container")
		So(strings.TrimSpace(result.Stdout()), ShouldEqual, "e2e-test-container")
		So(result.ExitCode, ShouldEqual, 0)
	})
}

func TestAttach(t *testing.T) {
	Convey("docker run --name attach -d redis", t, func(c C) {
		var resultRun *icmd.Result
		var resultAttach *icmd.Result
		var resultStop *icmd.Result
		done := make(chan bool)

		resultRun = icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "--name", "attach", "-d", "redis")

		go func() {
			resultAttach = icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "attach", "attach")
		}()

		go func() {
			resultStop = icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "container", "stop", "attach")
			time.Sleep(1 * time.Second)
			done <- true
		}()

		<-done

		So(resultAttach.ExitCode, ShouldEqual, 1)
		So(resultRun.ExitCode, ShouldEqual, 0)
		So(resultStop.ExitCode, ShouldEqual, 0)
		// So(resultAttach.Stdout(), ShouldEndWith, "# Redis is now ready to exit, bye bye...\n")
		resultStop = icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", "attach")
	})
}

func TestBuild(t *testing.T) {
	Convey("docker build -t image-teste-build ./buildtest", t, func() {
		resultBuild := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "build", "-t", "image-teste-build", "build/")
		So(resultBuild.ExitCode, ShouldEqual, 0)
		So(resultBuild.Stdout(), ShouldContainSubstring, "Successfully built")
		resultRun := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "-d", "image-teste-build")
		resultLogs := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "logs", strings.TrimSpace(resultRun.Stdout()))
		So(resultLogs.Stdout(), ShouldEqual, "GO-HORSE build command test\n")
		resultRm := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", strings.TrimSpace(resultRun.Stdout()))
		So(resultRm.ExitCode, ShouldEqual, 0)
	})
}

func TestCommit(t *testing.T) {
	Convey("docker commit", t, func() {
		resultRun := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "-d", "redis")
		So(resultRun.ExitCode, ShouldEqual, 0)
		containerID := strings.TrimSpace(resultRun.Stdout())
		resultExec := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "exec", containerID, "bash", "-c", "echo go-horse_commit_test > /test.test")
		So(resultExec.ExitCode, ShouldEqual, 0)
		resultCommit := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "commit", containerID, "commit_test")
		So(resultCommit.ExitCode, ShouldEqual, 0)
		resultRunCommit := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "-d", "commit_test")
		containerCommitID := strings.TrimSpace(resultRunCommit.Stdout())
		So(resultRunCommit.ExitCode, ShouldEqual, 0)
		resultExecCommit := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "exec", containerCommitID, "bash", "-c", "cat /test.test")
		So(resultExecCommit.ExitCode, ShouldEqual, 0)
		So(strings.TrimSpace(resultExecCommit.Stdout()), ShouldEqual, "go-horse_commit_test")
		resultRMs := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", containerCommitID, containerID)
		So(resultRMs.ExitCode, ShouldEqual, 0)
	})
}

func TestCP(t *testing.T) {
	Convey("docker cp", t, func() {
		resultRun := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "--name", "cp_test", "-d", "redis")
		So(resultRun.ExitCode, ShouldEqual, 0)
		resultCp := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "cp", "build/Dockerfile", "cp_test:/data")
		So(resultCp.ExitCode, ShouldEqual, 0)
		resultExec := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "exec", "cp_test", "ls")
		So(resultExec.ExitCode, ShouldEqual, 0)
		So(strings.TrimSpace(resultExec.Stdout()), ShouldEqual, "Dockerfile")
		resultRM := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", strings.TrimSpace(resultRun.Stdout()))
		So(resultRM.ExitCode, ShouldEqual, 0)
	})
}

func TestContainerStats(t *testing.T) {
	Convey("docker container stats", t, func() {
		resultRun := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "--name", "stats_container", "-d", "redis")
		So(resultRun.ExitCode, ShouldEqual, 0)

		var resultStats *icmd.Result
		done := make(chan string)
		go func() {
			command := icmd.Cmd{Command: append([]string{"docker", "-H", "tcp://localhost:7070", "stats", "stats_container"}), Timeout: 5 * time.Second}
			resultStats = icmd.StartCmd(command)
			time.Sleep(5 * time.Second)
			resultStats.Cmd.Process.Signal(os.Interrupt)
			done <- resultStats.Combined()
		}()

		msg := <-done

		So(msg, ShouldContainSubstring, "CONTAINER ID")
		So(msg, ShouldContainSubstring, "CPU %")
		So(msg, ShouldContainSubstring, "MEM USAGE / LIMIT")
		So(msg, ShouldContainSubstring, "NET I/O")
		So(msg, ShouldContainSubstring, "stats_container")

		So(resultStats.ExitCode, ShouldEqual, 0)

		resultRM := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", strings.TrimSpace(resultRun.Stdout()))
		So(resultRM.ExitCode, ShouldEqual, 0)
	})
}

// func TestStackDeploy(t *testing.T) {
// 	Convey("docker swarm all-in-one test suite", t, func() {
// 		resultSwarmInit := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "swarm", "init")
// 		So(resultSwarmInit.ExitCode, ShouldEqual, 0)
// 		So(strings.TrimSpace(resultSwarmInit.Stdout()), ShouldStartWith, "Swarm initialized")
// 		// fmt.Println("swarm init :: ", resultSwarmInit.Combined())

// 		resultStackDeploy := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "stack", "deploy", "--compose-file", "stack-deploy/docker-compose.yml", "stack-test")
// 		So(resultStackDeploy.ExitCode, ShouldEqual, 0)
// 		// fmt.Println("stack deploy :: ", resultStackDeploy.Combined())

// 		resultStackLs := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "stack", "ls")
// 		So(resultStackLs.ExitCode, ShouldEqual, 0)
// 		// fmt.Println("stack ls :: ", resultStackLs.Combined())

// 		resultServiceLs := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "service", "ls")
// 		So(resultServiceLs.ExitCode, ShouldEqual, 0)
// 		// fmt.Println("service ls :: ", resultServiceLs.Combined())

// 		var resultServiceLogs *icmd.Result

// 		done := make(chan string)
// 		go func() {
// 			command := icmd.Cmd{Command: append([]string{"docker", "-H", "tcp://localhost:7070", "service", "logs", "stack-test_redis"})}
// 			resultServiceLogs = icmd.StartCmd(command)
// 			time.Sleep(3 * time.Second)
// 			resultServiceLogs.Cmd.Process.Signal(os.Interrupt)
// 			done <- resultServiceLogs.Stdout()
// 		}()

// 		logs := <-done
// 		So(resultServiceLogs.ExitCode, ShouldEqual, 0)
// 		So(logs, ShouldContainSubstring, "Redis version")
// 		So(logs, ShouldContainSubstring, "Server initialized")
// 		So(logs, ShouldContainSubstring, "Ready to accept connections")
// 		// fmt.Println("service logs :: ", logs)

// 		resultStackRm := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "stack", "rm", "stack-test")
// 		So(resultStackRm.ExitCode, ShouldEqual, 0)
// 		// fmt.Println("stack rm :: ", resultStackRm.Combined())

// 		resultSwarmLeave := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "swarm", "leave", "--force")
// 		So(resultSwarmLeave.ExitCode, ShouldEqual, 0)
// 		So(strings.TrimSpace(resultSwarmLeave.Stdout()), ShouldEqual, "Node left the swarm.")
// 		// fmt.Println("swarm leave :: ", resultSwarmLeave.Combined())
// 	})

// }

func TestGoHorseNoFilters(t *testing.T) {
	Convey("go horse : no active filters", t, func() {
		removeContents(config.JsFiltersPath)
		time.Sleep(time.Second) // see list.go line 140
		response, err := http.Get("http://localhost:7070/active-filters")
		So(err, ShouldBeNil)
		So(response, ShouldNotBeNil)
		defer response.Body.Close()
		bodyBytes, er := ioutil.ReadAll(response.Body)
		So(er, ShouldBeNil)
		So(bodyBytes, ShouldNotBeNil)
		json := string(bodyBytes)
		requestFilters := gjson.Get(json, "request")
		responseFilters := gjson.Get(json, "response")
		resultPS := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "ps")
		So(requestFilters.Raw, ShouldBeIn, []string{"null", "[]"})
		So(responseFilters.Raw, ShouldBeIn, []string{"null", "[]"})
		So(resultPS.ExitCode, ShouldEqual, 0)
		So(strings.TrimSpace(resultPS.Stdout()), ShouldEndWith, "NAMES")
	})
}

func TestGoHorseOneResponseFilterRewritePsCommandBody(t *testing.T) {
	Convey("go horse : one active filter -> rewriting docker ps daemon response body", t, func() {
		er := removeContents(config.JsFiltersPath)
		So(er, ShouldBeNil)
		er = copy("jsFilters/000.response.rewrite_body_ps_command.js", config.JsFiltersPath+"/000.response.rewrite_body_ps_command.js")
		So(er, ShouldBeNil)
		time.Sleep(time.Second) // DirWatcher interval : 1 second loop
		response, err := http.Get("http://localhost:7070/active-filters")
		So(err, ShouldBeNil)
		So(response, ShouldNotBeNil)
		defer response.Body.Close()
		bodyBytes, er := ioutil.ReadAll(response.Body)
		So(er, ShouldBeNil)
		So(bodyBytes, ShouldNotBeNil)
		json := string(bodyBytes)
		requestFilters := gjson.Get(json, "request")
		responseFilters := gjson.Get(json, "response")
		resultPS := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "ps")
		So(requestFilters.Raw, ShouldBeIn, []string{"null", "[]"})
		So(responseFilters.Raw, ShouldStartWith, "[{\"Name\":\"rewrite_body_ps_command\",\"Order\":0,\"PathPattern\":\"containers/json")
		So(resultPS.ExitCode, ShouldEqual, 0)
		So(strings.TrimSpace(resultPS.Stdout()), ShouldContainSubstring, "go-horse")
		So(strings.TrimSpace(resultPS.Stdout()), ShouldContainSubstring, "go-horse.sh")
		So(strings.TrimSpace(resultPS.Stdout()), ShouldContainSubstring, "go-horse-image")
		So(strings.TrimSpace(resultPS.Stdout()), ShouldContainSubstring, "go-horse-name")
		So(strings.TrimSpace(resultPS.Stdout()), ShouldContainSubstring, "About a go-horse ago")
		removeContents(config.JsFiltersPath)
	})
}

func TestGoHorseFilterScope(t *testing.T) {
	Convey("go horse : 2 active filters -> one filter sets a variable and change a container label, another filter reads and verify that variable stored in request scope", t, func() {
		er := removeContents(config.JsFiltersPath)
		So(er, ShouldBeNil)
		er = copy("jsFilters/000.request.setVar.js", config.JsFiltersPath+"/000.request.setVar.js")
		er = copy("jsFilters/000.response.getVar.js", config.JsFiltersPath+"/000.response.getVar.js")
		time.Sleep(time.Second)

		resultRun := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "-d", "--label", "test_label=go-horse-label", "redis")
		So(resultRun.ExitCode, ShouldEqual, 0)
		containerID := strings.TrimSpace(resultRun.Stdout())

		resultInspect := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "inspect", containerID)
		So(resultInspect.ExitCode, ShouldEqual, 0)
		So(resultInspect.Stdout(), ShouldContainSubstring, "go-horse-label_edited_by_filter_set_var")

		resultRM := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", containerID)
		So(strings.TrimSpace(resultRM.Stdout()), ShouldEqual, containerID)
		So(resultRM.ExitCode, ShouldEqual, 0)
		removeContents(config.JsFiltersPath)
	})
}

func TestGoHorseGolangFilter(t *testing.T) {
	Convey("go horse : Golang filter", t, func() {

	})
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func copy(src string, dst string) error {
	data, er := ioutil.ReadFile(src)
	if er != nil {
		fmt.Println("copy : error reading file :> ", er.Error())
		return er
	}
	er = ioutil.WriteFile(dst, data, 0644)
	if er != nil {
		fmt.Println("copy : error writing file in dst :> ", er.Error())
		return er
	}
	return nil
}
