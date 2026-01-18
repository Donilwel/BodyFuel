build:
	@docker build --pull --build-arg GOPROXY=$${GOPROXY} --build-arg GOPRIVATE=$${GOPRIVATE} -t ${IMAGE}:${TAG} .

swag:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/main.go -o api-schemas
	@echo "Swagger documentation generated successfully"

