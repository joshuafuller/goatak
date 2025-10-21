FROM golang:1.24 AS builder

WORKDIR /build

COPY . .
RUN make build

FROM alpine:latest

EXPOSE 8080/tcp
EXPOSE 8088/tcp
EXPOSE 8089/tcp
EXPOSE 8443/tcp
EXPOSE 8446/tcp
EXPOSE 8999/tcp
EXPOSE 8999/udp

WORKDIR /app
RUN apk update && apk upgrade && apk add --no-cache bash openssl zip
COPY --from=builder /build/dist/goatak_server /app/goatak_server
COPY ./cert/*.sh /app/cert/
COPY ./data /app/data
COPY ./docker/run.sh /app/
COPY ./goatak_server.yml /app/
CMD ["./run.sh"]