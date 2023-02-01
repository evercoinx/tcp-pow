SERVER_BIN = tcpserver
CLIENT_BIN = tcpclient
IMAGE_NAME = tcp-pow-server
IMAGE_TAG = latest

deps:
	go mod tidy
	go mod vendor

server-build:
	go build -mod=vendor -o bin/$(SERVER_BIN) cmd/$(SERVER_BIN)/main.go

client-build:
	go build -mod=vendor -o bin/$(CLIENT_BIN) cmd/$(CLIENT_BIN)/main.go

server-up: server-build
	./bin/$(SERVER_BIN)

client-up: client-build
	./bin/$(CLIENT_BIN)

docker-server-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-server-run: docker-server-build
	docker run -d -p 8000:8000 --name $(IMAGE_NAME) --rm $(IMAGE_NAME):$(IMAGE_TAG)
