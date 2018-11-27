
testar comandos no proxy : 
docker -H=localhost:8080 ps -a

Entidades api v 1.37 :
docker run --rm -it -e GOPATH=$HOME/go:/go -v $HOME:$HOME -w $(pwd) quay.io/goswagger/swagger generate model -f swagger.yaml