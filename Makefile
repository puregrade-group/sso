run:
	go run ./cmd/main.go --config=./config/config.yaml

doc:
	godoc -http=localhost:6060 # http://localhost:6060/pkg/github.com/puregrade/puregrade-auth/?m=all

generate-proto:
	protoc --go_out=./internal/transport/grpc/proto --go-grpc_out=./internal/transport/grpc/proto ./internal/transport/grpc/proto/auth.proto

# up all migrations
mgrs-up:
	go run ./cmd/migrator/main.go --storage-path=storage/sso.db  --migrations-path=migrations

test-mgrs-up:
	go run ./cmd/migrator/main.go --storage-path="./storage/sso.db"  --migrations-path="./tests/migrations" --migrations-table="test"