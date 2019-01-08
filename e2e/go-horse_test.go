package tests

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/server"
	. "github.com/smartystreets/goconvey/convey"
	"gotest.tools/icmd"
)

func init() {
	config.SetLogLevel("error")
	config.SetPort(":7070")
	go func() {
		server.GoHorse()
	}()
	SetDefaultFailureMode(FailureContinues)
}

// https://github.com/docker/cli/blob/master/e2e/container/run_test.go
// https://github.com/smartystreets/goconvey/
// https://github.com/ory/dockertest

func TestRun(t *testing.T) {
	Convey("docker run --name e2e-test-container -d redis", t, func() {
		result := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "--name", "e2e-test-container", "-d", "redis")
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
		fmt.Printf(">>>>>>>>>>>>>>>>>>> %s || %s", resultRun.Stdout(), resultRun.Stderr())
		So(resultRun.ExitCode, ShouldEqual, 0)
		resultCp := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "cp", "build/Dockerfile", "cp_test:/data")
		So(resultCp.ExitCode, ShouldEqual, 0)
		resultExec := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "exec", "cp_test", "ls")
		fmt.Printf(">>>>>>>>>>>>>>>>>>> %s || %s", resultExec.Stdout(), resultExec.Stderr())
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
