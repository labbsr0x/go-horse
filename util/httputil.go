package util

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/websocket"
)

// GetRequestParameter GetRequestParameter
func GetRequestParameter(formValues map[string][]string, param string) (value string) {
	values, ok := formValues[param]
	if ok {
		value = values[0]
	}
	return
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
