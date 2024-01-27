# Single Sign-On (SSO)

### langs
- [ru](https://github.com/puregrade-group/sso/blob/master/README_ru.md)

### Description:
A demonstration project created to showcase my skills.

The project itself is an implementation of the Single Sign-On technology. It enables users to register/login in one service and use their unique token and roles/permissions across all registered applications.

#### Technology Stack: Go, gRPC, SQLite, Docker

### Project Structure:
```
├───cmd
│   └───migrator // Wrapper for migrate
├───config // Configurations
├───internal
│   ├───app // Application components integration
│   │   └───grpc // gRPC server application code
│   ├───config // Application config structure
│   ├───domain
│   │   └───models // Shared structures
│   ├───service // Service layer
│   │   ├───acs
│   │   └───auth
│   ├───storage // Data storage layer
│   │   └───sqlite
│   ├───transport // Data transport layer
│   │   └───grpc
│   │       ├───acs // Files for working with roles/permissions
│   │       └───auth // Files for user registration/login
│   └───utils // Useful components reused in the application
│       ├───jwt
│       └───slogpretty
├───migrations // Migration files
├───storage // Database files
└───tests // Tests
├───migrations
└───suite
```

### Installation, Building, and Running:

#### Prerequisites:
1. go compiler 1.21.5
2. git
3. Docker
4. make
5. Postman

#### Steps:
1. Clone the repository: `git clone https://github.com/puregrade-group/sso ./my/favorite/dir`
2. Install dependencies: `go mod download`
3. Create the database and populate tables: `make mgrs-up` or `go run ./cmd/migrator/main.go --storage-path=storage/sso.db  --migrations-path=migrations`
4. For testing, populate the necessary test data: `make test-mgrs-up` or `go run ./cmd/migrator/main.go --storage-path="./storage/sso.db"  --migrations-path="./tests/migrations" --migrations-table="test"`
5. Run the application: `make run` or `go run ./cmd/main.go --config=./config/config.yaml`
6. To test the functionality, you can run the tests using `go test`, send requests through Postman, or write your own client for this application. To do this, you will need to refer to https://github.com/puregrade-group/protos and find the .proto files there for Postman or import the latest version of the generated files from this repository for your own client.

or

5. Build the Docker image: `docker build --tag image-name .`
6. Run the container: `docker run -p 50051:50051/tcp --name container-name <image_id>`
7. Test the functionality.

##### Examples:

Application logs upon startup:
<p align="left"><img width="400px" src="https://github.com/puregrade-group/sso/raw/master/example/execute_log.png" alt="execute_log.png"/></p>

Output from Postman:
<p align="left"><img width="400px" src="https://github.com/puregrade-group/sso/raw/master/example/postman_output.png" alt="postman_output.png"/></p>

Output from tests:
<p align="left"><img width="400px" src="https://github.com/puregrade-group/sso/raw/master/example/test_output.png" alt="test_output.png"/></p>