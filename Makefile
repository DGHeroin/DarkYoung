## 默认情况下不需要支持
all:
	@echo "You don't to run this"
proto:
	protoc --proto_path=$(GOPATH)/src:. --go_out=plugins=grpc:. ./proto/message.proto