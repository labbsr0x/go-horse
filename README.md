## **GO-HORSE** : DOCKER DAEMON PROXY/FILTER

>The software in the middle the communication between docker client and daemon, allowing you to intercept all commands and, by example, do access control or add tags in a container during its creation, change its name, alter network definition, redifine volumes, rewrite the whole body if you want, and so on. Take control. Do what you need.

### How it works

Docker (http) commands sent from the client to the deamon are intercepted by creating filters in go-horse. This filters can be implemented either in JavaScript or Golang. You should inform a *path pattern* to match a command url (check [api documentation](https://docs.docker.com/engine/api/v1.39/)), a *invoke* property telling if you want the filter to run at the Request time, before the request hit the daemon, or on Response time, after the daemon has processed the request. Once your filter gets a request, you have all the means to implement the rules that your business needs. Rewrite a url to the docker daemon? Check the user identity in another system? Send a http request and break the filter chain based on the response? Add metadata to a container? Change container properties? Compute specific metrics?  Blacklist some commands? Ok, can do this, and many more.

### Running

``` docker-compose
version: '3.7'
services:
  proxy:
    image: labbs/go-horse
    network_mode: bridge
    ports: 
      - 8080:8080
    environment: 
      - DOCKER_HOST=/var/run/docker.sock
      - DOCKER_SOCK=unix:///var/run/docker.sock
      - TARGET_HOSTNAME=http://localhost
      - LOG_LEVEL=debug
      - PRETTY_LOG=true
      - PORT=:8080
      - JS_FILTERS_PATH=/app/go-horse/filters
      - GO_PLUGINS_PATH=/app/go-horse/plugins
    volumes: 
      - /var/run/docker.sock:/var/run/docker.sock
      - /home/bruno/go-horse:/app/go-horse
```
Set the environment variable `DOCKER_HOST` to `tcp://localhost:8080` or test a single command adding -H attribute : `docker -H=localhost:8080 ps -a` and watch the go-horse container logs

### Filtering requests using JavaScript
According to the environment variable `JS_FILTERS_PATH` you have to place you JavaScript filters there to get them loaded in the go-horse filter chain. These file's name have to obey a pattern :

#### 000.request.test.js => {order}.{invoke}.{name}.{extension}  
| Property  | Values allowed | Description|
| ------------- | ------------- |------------|
| Order  | [0-9]{1,3} | Filter execution order are sorted by this property and should be unique.| 
| Invoke  | *Request* or *Response* | Filter will be invoked *before*(Request) or *after*(Response) the command are sent to daemon|
| Name | (.*?) | A name for your filter |


### Filtering requests using Go

### JS versus GO - information to help your choice
benchmark