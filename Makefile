SERVER_BIN = tcpserver
CLIENT_BIN = tcpclient

deps:
	go mod tidy
	go mod vendor

server-build:
	go build -mod=vendor -o bin/$(SERVER_BIN) cmd/$(SERVER_BIN)/main.go

server-run: server-build
	./bin/$(SERVER_BIN)

client-build:
	go build -mod=vendor -o bin/$(CLIENT_BIN) cmd/$(CLIENT_BIN)/main.go

client-run: client-build
	./bin/$(CLIENT_BIN)

service-up:
	docker-compose up -d

service-down:
	docker-compose down
