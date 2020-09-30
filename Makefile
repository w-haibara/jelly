bot_deploy: *.go */*.go
	gofmt -w *.go */*.go
	go build -o bot_deploy main.go

.PHONY: init
init:
	go mod init jelly

