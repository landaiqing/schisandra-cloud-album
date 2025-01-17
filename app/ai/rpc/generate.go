package rpc

//go:generate goctl rpc protoc ai.proto --go_out=. --go-grpc_out=. --zrpc_out=. --client=true -m --style=go_zero
