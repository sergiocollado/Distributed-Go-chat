# Distributed-Go-chat

reference: https://www.freecodecamp.org/news/how-to-build-a-production-grade-distributed-chatroom-in-go-full-handbook/

## How to run the server and the client

To start the server: 

```bash
go run ./cmd/server
```

To start the client: 

```bash
go run ./cmd/client
```

## Network inspection

In a local machine you can check it with: `sudo tcpdump -A -i lo tcp port 9000`.

## How to run the server as a daemon

To run the server as a daemon, create: /etc/systemd/system/chatroom.service:    

```
[Unit]
Description=Chatroom Server
After=network.target

[Service]
Type=simple
User=chatroom
WorkingDirectory=/opt/chatroom
ExecStart=/opt/chatroom/server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

To build the server: `go build -o server cmd/server/main.go`

copy it into the deployment location: 

```bash
sudo mkdir -p /opt/chatroom
sudo cp server /opt/chatroom/
sudo mkdir -p /opt/chatroom/chatdata
```

Create a dedicated user for running the service:


```bash
sudo useradd -r -s /bin/false chatroom
sudo chown -R chatroom:chatroom /opt/chatroom
```

Enable and start the service:

```bash
sudo systemctl enable chatroom
sudo systemctl start chatroom
```

Check that it's running:

```bash
sudo systemctl status chatroom
```

You can view logs with:

```bash
sudo journalctl -u chatroom -f
```

## How to deploy it with docker

Create a docker file: 

```
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/chatdata ./chatdata
EXPOSE 9000
CMD ["./server"]

```

This uses a multi-stage build. The first stage (builder) uses the full Go image to compile your server. The second stage uses a minimal Alpine Linux image and copies only the compiled binary. This keeps the final image small (about 20MB instead of 800MB).

To run it: 

```
docker build -t chatroom .
docker run -p 9000:9000 -v $(pwd)/chatdata:/root/chatdata chatroom
```

The -p 9000:9000 maps port 9000 in the container to port 9000 on your host, making the chatroom accessible. The -v $(pwd)/chatdata:/root/chatdata mounts your local chatdata directory into the container, so messages persist even if you stop and remove the container.

For running it in Docker compose or K8, use the config: 


```
version: '3.8'
services:
  chatroom:
    build: .
    ports:
      - "9000:9000"
    volumes:
      - ./chatdata:/root/chatdata
    restart: unless-stopped
```

and run it with: 'docker-compose up -d'