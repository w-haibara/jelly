deploy_bot: *.go */*.go
	gofmt -w *.go */*.go
	go build -o deploy_bot main.go

.PHONY: init
init:
	go mod init deploy_bot

