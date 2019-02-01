FROM golang:latest 

ENV p /go/src/gitex.labbs.com.br/labbsr0x/proxy/go-horse

RUN mkdir -p ${p}
ADD . ${p}
WORKDIR ${p}
RUN go get -v ./...

# TODO try
# && BUILDTIME=$(date --utc --rfc-3339 ns 2> /dev/null | sed -e 's/ /T/')\

RUN GIT_COMMIT=$(git rev-parse --short HEAD 2> /dev/null || true) \
 && BUILDTIME=$(TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ') \
 && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
    -X gitex.labbs.com.br/labbsr0x/proxy/go-horse/config.GitCommit=${GIT_COMMIT} \
    -X gitex.labbs.com.br/labbsr0x/proxy/go-horse/config.BuildTime=${BUILDTIME}" \
    -a -installsuffix cgo -o /main

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /main /
CMD ["/main"]