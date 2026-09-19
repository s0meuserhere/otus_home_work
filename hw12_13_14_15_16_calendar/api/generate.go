package api

//go:generate sh -c "protoc -I. -I\"$(dirname \"$(command -v protoc)\")/../include\" --go_out=../internal/server/grpc/gen --go_opt=paths=source_relative --go-grpc_out=../internal/server/grpc/gen --go-grpc_opt=paths=source_relative event_service.proto"
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 -generate types,std-http,spec,skip-prune -package gen -o ../internal/server/http/gen/event_service.gen.go event_service.yaml
