run:
	go run ./...

build-windows:
	GOOS=windows GOARCH=386 go build -o ./upsync.exe ./cmd/upsync/upsync.go 

build-linux:
	GOOS=linux GOARCH=amd64 go build -o ./upsync ./cmd/upsync/upsync.go