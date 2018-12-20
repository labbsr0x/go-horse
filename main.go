package main

import (
	"os"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/handlers"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	app := iris.New()
	app.Use(recover.New())
	app.Any("*", handlers.ProxyHandler)
	app.Run(iris.Addr(config.Port))
	// os.Setenv("DOCKER_API_VERSION", "1.39")
	// cli, err := client.NewEnvClient()
	// if err != nil {
	// 	panic(err)
	// }

	// containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	// if err != nil {
	// 	panic(err)
	// }

	// for _, container := range containers {
	// 	fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	// }
	// cliflect := reflect.ValueOf(&cli)
	// fmt.Println(cliflect.NumMethod())
	// post := reflect.ValueOf(&cli).MethodByName("postRaw")
	// private.SetAccessible(post)

	// var ctx context.Context

	// params := make([]reflect.Value, 5)
	// params[0] = reflect.ValueOf(ctx)

	// // ctx, "/v1.39/containers/d2c491d0a221/wait?condition=next-exit"

	// // ctx context.Context, path string, query url.Values, body io.Reader, headers map[string][]string) (serverResponse, error
	// response := post.Call(params) // stdout map[k:[v]]
	// for val := range response {
	// 	fmt.Println(fmt.Sprintf("%#v", val))
	// }

}

func init() {
	// zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(config.LogLevel)
	if config.PrettyLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// panic (zerolog.PanicLevel, 5)
	// fatal (zerolog.FatalLevel, 4)
	// error (zerolog.ErrorLevel, 3)
	// warn (zerolog.WarnLevel, 2)
	// info (zerolog.InfoLevel, 1)
	// debug (zerolog.DebugLevel, 0)

}

// conn, err := net.DialTimeout("unix", "/var/run/docker.sock", time.Duration(10*time.Second))
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	config, err := websocket.NewConfig(
// 		"/containers/7f4c141f99a9f0ab8396627de3ca4817064868680ba7a46ab51c7c4da5f8db4f/attach/ws?stream=1&stdin=1&stdout=1&stderr=1",
// 		"http://localhost",
// 	)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	ws, err := websocket.NewClient(config, conn)
// 	defer ws.Close()

// 	// expected := []byte("hello")
// 	// actual := make([]byte, len(expected))

// 	// outChan := make(chan error)
// 	// go func() {
// 	// 	_, err := io.ReadFull(ws, actual)
// 	// 	outChan <- err
// 	// 	close(outChan)
// 	// }()

// 	// inChan := make(chan error)
// 	// go func() {
// 	// 	_, err := ws.Write(expected)
// 	// 	inChan <- err
// 	// 	close(inChan)
// 	// }()

// 	// select {
// 	// case err := <-inChan:
// 	// 	fmt.Println(err)
// 	// case <-time.After(5 * time.Second):
// 	// 	fmt.Println("TIMEOUT")
// 	// }

// 	// select {
// 	// case err := <-outChan:
// 	// 	fmt.Println(err)
// 	// case <-time.After(5 * time.Second):
// 	// 	fmt.Println("TIMEOUT")
// 	// }

// 	for err == nil {
// 		var message string
// 		err = websocket.Message.Receive(ws, &message)
// 		// if err != nil {
// 		// 	fmt.Printf("Error::: %s\n", err.Error())
// 		// 	break
// 		// }
// 		if len(message) > 0 {
// 			fmt.Println("message : " + message)
// 		}
// 	}
