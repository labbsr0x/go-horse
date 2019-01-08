package tests

import (
	"strings"
	"testing"
	"time"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/server"
	. "github.com/smartystreets/goconvey/convey"
	"gotest.tools/icmd"
)

const containerName = "e2e-test-container"

func init() {
	config.SetLogLevel("error")
	config.SetPort(":7070")
	go func() {
		server.GoHorse()
	}()
}

// https://github.com/docker/cli/blob/master/e2e/container/run_test.go
// https://github.com/smartystreets/goconvey/
// https://github.com/ory/dockertest

func TestRun(t *testing.T) {
	Convey("docker run --name e2e-test-container -d redis", t, func() {
		result := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "--name", containerName, "-d", "redis")
		So(len(strings.TrimSpace(result.Stdout())), ShouldEqual, 64)
		So(result.ExitCode, ShouldEqual, 0)
	})
}

func TestBuild(t *testing.T) {
	Convey("docker build -t image-teste-build ./buildtest", t, func() {
		resultBuild := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "build", "-t", "image-teste-build", "build/")
		So(resultBuild.ExitCode, ShouldEqual, 0)
		So(resultBuild.Stdout(), ShouldContainSubstring, "Successfully built")
		Convey("Create a container from image-teste-build image", func() {
			resultRun := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "run", "-d", "image-teste-build")
			resultLogs := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "logs", strings.TrimSpace(resultRun.Stdout()))
			So(resultLogs.Stdout(), ShouldEqual, "GO-HORSE build command test\n")
			resultRm := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", strings.TrimSpace(resultRun.Stdout()))
			So(resultRm.ExitCode, ShouldEqual, 0)
		})
	})
}

func TestAttach(t *testing.T) {
	// Convey("Doing some asyn testing", t, func(c C) {
	//     done := make(chan bool)
	//     go func() {
	//         c.So(2, ShouldEqual, 2)
	//         done <- true
	//     }()
	//     _ = <-done
	// })
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

func TestRemove(t *testing.T) {
	Convey("docker run -f e2e-test-container", t, func() {
		result := icmd.RunCommand("docker", "-H", "tcp://localhost:7070", "rm", "-f", containerName)
		So(strings.TrimSpace(result.Stdout()), ShouldEqual, containerName)
		So(result.ExitCode, ShouldEqual, 0)
	})
}
