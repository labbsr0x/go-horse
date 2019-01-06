package tests

import (
	"context"
	"testing"

	"gitex.labbs.com.br/labbsr0x/proxy/go-horse/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	. "github.com/smartystreets/goconvey/convey"
)

var dockerCliTest *client.Client

func init() {
	var err error
	dockerCliTest, err = client.NewClientWithOpts(client.WithVersion(config.DockerAPIVersion), client.WithHost("tcp://localhost:7070"))
	if err != nil {
		panic(err)
	}
}

// func TestSomething(t *testing.T) {
// 	Convey("1 should equal 1", t, func() {
// 		So(1, ShouldEqual, 1)
// 	})
// }

// func TestAnotherTest(t *testing.T) {
// 	Convey("Comparing two variables", t, func() {
// 		myVar := "Hello, world!"

// 		Convey(`"Asdf" should NOT equal "qwerty"`, func() {
// 			So("Asdf", ShouldNotEqual, "qwerty")
// 		})

// 		Convey("myVar should not be nil", func() {
// 			So(myVar, ShouldNotBeNil)
// 		})
// 	})
// }

// func TestSomething2(t *testing.T) {
// 	Convey("2 should not equal 3", t, func() {
// 		So(2, ShouldNotEqual, 3)
// 	})
// }

// https://github.com/docker/cli/blob/master/e2e/container/run_test.go

func TestListContainers(t *testing.T) {
	ctx := context.Background()
	options := types.ContainerListOptions{All: true}
	var containers []types.Container
	var err error
	Convey("When i exec ContainerList method in docker client", t, func() {
		containers, err = dockerCliTest.ContainerList(ctx, options)
		So(err, ShouldBeNil)
		Convey("Results should be greater then 0", func() {
			So(len(containers), ShouldBeGreaterThan, 0)
		})
		Convey("Go-horse container should be running", func() {
			found := false
			for _, container := range containers {
				if container.Image == "go-horse" {
					found = true
				}
			}
			So(found, ShouldBeTrue)
		})
	})
}
