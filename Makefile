SRVBIN = tcpserver
CLIBIN = tcpclient

deps:
	go mod vendor

server-build:
	go build -mod=vendor -o bin/$(SRVBIN) cmd/$(SRVBIN)/main.go

client-build:
	go build -mod=vendor -o bin/$(CLIBIN) cmd/$(CLIBIN)/main.go

server-up: server-build
	./bin/$(SRVBIN)

client-up: client-build
	./bin/$(CLIBIN)
