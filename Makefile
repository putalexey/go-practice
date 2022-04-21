proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/app/proto/grpc.proto

sort_imports:
	goimports -w -local github.com/putalexey ./cmd/ ./internal/

build:
	cd cmd/shortener/; go build -o shortener
#docs:
#	cd internal/app/shortener/
#	swag init -g ./shortner.go
