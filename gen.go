package ebully

//go:generate protoc --proto_path=./proto --go_out=./node --go_opt=paths=source_relative --go-grpc_out=./node --go-grpc_opt=paths=source_relative ./proto/node.proto
