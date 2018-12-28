FROM golang:latest 

ENV p /go/src/gitex.labbs.com.br/labbsr0x/proxy/go-horse

RUN mkdir -p ${p}
ADD . ${p}
WORKDIR ${p}
RUN go get -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /main .
CMD ["/app/main"]

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /main /
CMD ["/main"]