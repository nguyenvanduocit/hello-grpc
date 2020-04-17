generate:
	protoc -I proto/ --go_out=plugins=grpc:proto/productservice proto/**/*.proto
