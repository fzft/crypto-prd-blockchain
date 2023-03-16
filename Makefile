build:
	go build -o bin/main


run: build
	./bin/main


test:
	go test ./...


proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto


build-linux:
	GOOS=linux GOARCH=arm64 go build -o bin/bloker-linux-arm64

scp: build-linux
	scp bin/bloker-linux-arm64 mos@192.168.64.8:/home/data/

.PHONY: proto