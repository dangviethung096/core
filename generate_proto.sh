echo "Generating Go files from protobuf files for core"

# Clean existing generated files
rm -rf *.pb.go

# Generate Go files
protoc --go_out=./ --go_opt=paths=source_relative --go-grpc_out=./ --go-grpc_opt=paths=source_relative http_model.proto