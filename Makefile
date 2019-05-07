all:
	protoc --proto_path=$(GOPATH)/src:. --go_out=plugins=grpc:. ./proto/message.proto