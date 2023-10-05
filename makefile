build:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./func/handler ./cmd/main.go