package sockclient

import (
	"net"
	"net/http"
	"net/url"

	"github.com/rs/zerolog/log"
)

// Get lero lero
func Get(u string) *http.Client {
	url, err := url.Parse(u)
	if err != nil {
		log.Debug().Str("URL", url.RequestURI()).Err(err)
		return nil
	}
	transport := &http.Transport{}
	transport.DisableKeepAlives = true
	path := url.Path
	transport.Dial = func(proto, addr string) (net.Conn, error) {
		return net.Dial("unix", path)
	}
	url.Scheme = "http"
	url.Host = "unix-socket"
	url.Path = ""
	return &http.Client{Transport: transport}
}
