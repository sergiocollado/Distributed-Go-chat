# Distributed-Go-chat

reference: https://www.freecodecamp.org/news/how-to-build-a-production-grade-distributed-chatroom-in-go-full-handbook/

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

```
sudo mkdir -p /opt/chatroom
sudo cp server /opt/chatroom/
sudo mkdir -p /opt/chatroom/chatdata
```

Create a dedicated user for running the service:


```
sudo useradd -r -s /bin/false chatroom
sudo chown -R chatroom:chatroom /opt/chatroom
```

Enable and start the service:

```
sudo systemctl enable chatroom
sudo systemctl start chatroom
```

Check that it's running:

```
sudo systemctl status chatroom
```

You can view logs with:

```
sudo journalctl -u chatroom -f
```