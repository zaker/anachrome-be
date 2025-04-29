FROM --platform=linux/amd64 golang:1.24-alpine as builder
RUN apk update && \
    apk upgrade &&\
    apk add --no-cache ca-certificates git gcc musl-dev
RUN update-ca-certificates

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build --race -buildvcs=false
RUN go test --race ./...
RUN CGO_ENABLED=0\
    GOOS=linux\
    GOARCH=amd64 \
    go build \
        -ldflags="-w -s" \
        -o /bin/a-be \
        -buildvcs=false

FROM scratch as deployer
COPY --from=builder /bin/a-be /bin/a-be
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080 
ENTRYPOINT ["a-be","serve"]
