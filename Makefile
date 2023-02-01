SERVER_BIN = tcpserver
CLIENT_BIN = tcpclient
SERVER_IMAGE_TAG = go-tcp-pow/server:latest

deps:
	go mod tidy
	go mod vendor

server-build:
	go build -mod=vendor -o bin/$(SERVER_BIN) cmd/$(SERVER_BIN)/main.go

server-run: server-build
	./bin/$(SERVER_BIN)

docker-server-build:
	docker build -t $(SERVER_IMAGE_TAG) -f Dockerfile.server .

docker-server-run: docker-server-build
	docker run -d -p 8000:8000 --name go-go-tcp-pow --rm $(SERVER_IMAGE_TAG)

client-build:
	go build -mod=vendor -o bin/$(CLIENT_BIN) cmd/$(CLIENT_BIN)/main.go

client-run: client-build
	./bin/$(CLIENT_BIN)
