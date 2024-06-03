run:
	go run ./cmd/main.go --config=./config/config.yaml

doc:
	godoc -http=localhost:6060 # http://localhost:6060/pkg/github.com/puregrade/puregrade-auth/?m=all

# up all migrations
mgrs-up:
	go run ./cmd/migrator/main.go --storage-path=storage/sso.db  --migrations-path=migrations

test-mgrs-up:
	go run ./cmd/migrator/main.go --storage-path="./storage/sso.db"  --migrations-path="./tests/migrations" --migrations-table="test"

gen-auth:
	protoc -I pkg/protos/proto pkg/protos/proto/auth/auth.proto \
		--go_out=./pkg/protos/gen/go/ \
		--go_opt=paths=source_relative \
		--go-grpc_out=./pkg/protos/gen/go/ \
		--go-grpc_opt=paths=source_relative

gen-acs:
	protoc -I pkg/protos/proto pkg/protos/proto/acs/permissions.proto pkg/protos/proto/acs/roles.proto \
		--go_out=./pkg/protos/gen/go/ \
		--go_opt=paths=source_relative \
		--go-grpc_out=./pkg/protos/gen/go/ \
		--go-grpc_opt=paths=source_relative

gen-profile:
	protoc -I pkg/protos/proto pkg/protos/proto/profile/profile.proto \
		--go_out=./pkg/protos/gen/go/ \
		--go_opt=paths=source_relative \
		--go-grpc_out=./pkg/protos/gen/go/ \
		--go-grpc_opt=paths=source_relative

gen-protos:	gen-auth gen-acs gen-profile