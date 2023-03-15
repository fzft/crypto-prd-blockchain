build:
	go build -o bin/main


run: build
	./bin/main


test:
	go test ./...


proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto


.PHONY: proto