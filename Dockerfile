# Start from the latest golang base image
FROM abilioesteves/gowebbuilder:1.2.0

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Expose port 8080 to the outside world
EXPOSE 8080

RUN GIT_COMMIT=$(git rev-parse --short HEAD 2> /dev/null || true) \
 && BUILDTIME=$(TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ') \
 && CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w \
    -X gitex.labbs.com.br/labbsr0x/proxy/go-horse/config.GitCommit=${GIT_COMMIT} \
    -X gitex.labbs.com.br/labbsr0x/proxy/go-horse/config.BuildTime=${BUILDTIME}" \
    -a -installsuffix cgo -o /main

FROM scratch
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /main /

# Command to run the executable
CMD ["./main"]