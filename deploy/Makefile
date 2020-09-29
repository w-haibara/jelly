deploy_bot: *.go
	gofmt -w *.go
	go build -o deploy_bot *.go

.PHONY: init
init:
	go mod init deploy_bot

