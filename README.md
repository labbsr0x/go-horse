# GO-HORSE : DOCKER DAEMON PROXY/FILTER

>The software in the middle the communication between docker client and daemon, allowing you to intercept all commands and, by example, do access control or add tags in a container during its creation, change its name, alter network definition, redifine volumes and so on. Take control, do what you need.

## Running

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

## Filtering requests using JavaScript


## Filtering requests using Go

## JS versus GO - information to help your choice
benchmark