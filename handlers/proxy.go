package handlers

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/config"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/filters/list"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/model"
	sockclient "gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/sockClient"
	"gitex.labbs.com.br/labbsr0x/sandman-acl-proxy/util"
	"github.com/docker/cli/cli"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/kataras/iris"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/websocket"
)

var sockClient = sockclient.Get(config.DockerSockURL)
var dockerCli *client.Client

var waitChannel = make(chan int)

func init() {
	os.Setenv("DOCKER_API_VERSION", "1.39")
	os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")
	var err error
	dockerCli, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
}

// ProxyHandler lero-lero
func ProxyHandler(ctx iris.Context) {

	log.Debug().Str("request", ctx.String()).Msg("Receiving request")

	util.SetEnvVars(ctx)

	if ctx.Request().Body != nil {
		requestBody, erro := ioutil.ReadAll(ctx.Request().Body)
		if erro != nil {
			log.Error().Str("request", ctx.String()).Err(erro)
		}
		ctx.Values().Set("requestBody", string(requestBody))
	}

	filterReturn := before(ctx)

	if filterReturn.Err != nil {
		if filterReturn.Status == 0 {
			filterReturn.Status = http.StatusInternalServerError
		}
		ctx.StatusCode(filterReturn.Status)
		ctx.ContentType("application/json")
		ctx.WriteString(filterReturn.Err.Error())
		return
	}

	targetURL := ctx.Values().GetString("targetEndpoint")
	if targetURL == "" {
		targetURL = ctx.Request().URL.RequestURI()
	}

	request, newRequestError := http.NewRequest(ctx.Request().Method, config.TargetHostname+targetURL, strings.NewReader(ctx.Values().GetString("requestBody")))

	if newRequestError != nil {
		log.Error().Str("request", ctx.String()).Err(newRequestError).Msg("Error creating a new request in main handler")
	}

	for key, value := range ctx.Request().Header {
		request.Header[key] = value
	}

	log.Debug().Msg("Executing request for URL : " + targetURL + " ...")

	response, erre := sockClient.Do(request)

	if strings.Contains(targetURL, "attach") {

		context := context.Background()
		options := types.ContainerAttachOptions{}
		options.Stdout = true
		options.Stderr = true
		options.Stream = true
		resp, err := dockerCli.ContainerAttach(context, "teste", options)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		msgs := make(chan []byte)
		msgsErr := make(chan error)

		go func() {
			for {
				msg, er := resp.Reader.ReadBytes('\n')
				if er != nil {
					msgsErr <- er
					return
				}
				msgs <- msg
			}
		}()

		_, upgrade := ctx.Request().Header["Upgrade"]

		conn, _, err := ctx.ResponseWriter().Hijack()
		if err != nil {
			fmt.Println("ERRO >>>>>>>>>>>>>> ", err)
		}

		conn.Write([]byte{})

		if upgrade {
			fmt.Fprintf(conn, "HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n")
		} else {
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n")
		}

		go func() {
			// attachWs(func(message string) {
			// 	fmt.Fprintf(conn, "%b", message)
			// })
			for {
				select {
				case msg := <-msgs:
					fmt.Fprintf(conn, "%s", msg)
				case errr := <-msgsErr:
					fmt.Println("errrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrrr ", errr)
					conn.Close()
					return
				}
			}
		}()

		return

	}

	if strings.Contains(targetURL, "wait") {
		context := context.Background()
		resp, err := dockerCli.ContainerWait(context, "teste", container.WaitConditionNextExit)

		var respostaWait container.ContainerWaitOKBody
		var erroWait error = nil
		finish := false

		go func() {
			select {
			case result := <-resp:
				if result.Error != nil {
					respostaWait = result
					finish = true
				}
				if result.StatusCode != 0 {
					fmt.Println(cli.StatusError{StatusCode: int(result.StatusCode)})
					finish = true
				}
			case err0 := <-err:
				erroWait = err0
			}
		}()

		ctx.ContentType("application/json")
		ctx.Header("Transfer-Encoding", "chunked")

		ctx.StreamWriter(func(w io.Writer) bool {
			if finish {
				fmt.Fprintf(w, "%#v", respostaWait)
				return false
			}
			return true
		})

		return

	}

	if erre != nil {
		log.Error().Str("request", ctx.String()).Err(erre).Msg("Error executing the request in main handler")
		ctx.Next()
	}

	responseBody, erro := ioutil.ReadAll(response.Body)
	if erro != nil {
		ctx.WriteString("Erro parsear a resposta do token - " + erro.Error())
		log.Error().Str("request", ctx.String()).Err(erro).Msg("Error parsing response body in main handler")
	}

	for key, value := range response.Header {
		ctx.Header(key, value[0])
	}

	ctx.Values().Set("responseBody", string(responseBody))

	after(ctx)

	ctx.ContentType("application/json")
	ctx.StatusCode(response.StatusCode)
	ctx.WriteString(ctx.Values().GetString("responseBody"))

}

func before(ctx iris.Context) model.FilterReturn {
	requestPath := ctx.Path()
	log.Debug().Msg("Before the mainHandler: " + requestPath)

	var result model.FilterReturn
	for _, filter := range list.BeforeFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config()))
			result = filter.Exec(ctx, ctx.Values().GetString("requestBody"))
			log.Debug().Str("Filter output : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", result))
			if result.Operation == model.Write {
				log.Debug().Str("Body rewrite for filter - ", filter.Config().Name)
				ctx.Values().Set("requestBody", result.Body)
			}
			if !result.Next {
				log.Debug().Str("Filter chain canceled by filter - ", filter.Config().Name).Msg("lero-lero")
				break
			}
		}
	}
	return result
}

func after(ctx iris.Context) {
	requestPath := ctx.Path()
	log.Debug().Msg("After the mainHandler:" + requestPath)

	for _, filter := range list.AfterFilters() {
		if filter.MatchURL(ctx) {
			log.Debug().Str("Filter matched : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", filter.Config()))
			result := filter.Exec(ctx, ctx.Values().GetString("responseBody"))
			log.Debug().Str("Filter output : ", ctx.String()).Str("filter_config", fmt.Sprintf("%#v", result))
			if result.Operation == model.Write {
				log.Debug().Str("Body rewrite for filter - ", filter.Config().Name)
				ctx.Values().Set("responseBody", result.Body)
			}
			if !result.Next {
				log.Debug().Str("Filter chain canceled by filter - ", filter.Config().Name)
				break
			}
		}
	}
}

func attachWs(cbMessage func(msg string)) {
	conn, err := net.DialTimeout("unix", "/var/run/docker.sock", time.Duration(10*time.Second))
	if err != nil {
		fmt.Println(err)
	}
	config, err := websocket.NewConfig(
		"/containers/teste/attach/ws?stream=1&stdin=1&stdout=1&stderr=1",
		"http://localhost",
	)
	if err != nil {
		fmt.Println(err)
	}
	ws, err := websocket.NewClient(config, conn)

	if err != nil {
		fmt.Println(err)
	}

	defer ws.Close()

	// expected := []byte("hello")
	// actual := make([]byte, len(expected))

	// outChan := make(chan error)
	// go func() {
	// 	_, err := io.ReadFull(ws, actual)
	// 	outChan <- err
	// 	close(outChan)
	// }()

	// inChan := make(chan error)
	// go func() {
	// 	_, err := ws.Write(expected)
	// 	inChan <- err
	// 	close(inChan)
	// }()

	// select {
	// case err := <-inChan:
	// 	fmt.Println(err)
	// case <-time.After(5 * time.Second):
	// 	fmt.Println("TIMEOUT")
	// }

	// select {
	// case err := <-outChan:
	// 	fmt.Println(err)
	// case <-time.After(5 * time.Second):
	// 	fmt.Println("TIMEOUT")
	// }

	for err == nil {
		var message string
		err = websocket.Message.Receive(ws, &message)
		if err != nil {
			fmt.Printf("Error:::WEBSOCKET ATTACH >:>> %s\n", err.Error())
			break
		}
		if len(message) > 0 {
			cbMessage(message)
		}
	}
}
