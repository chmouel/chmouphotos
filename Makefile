.PHONY: vendor
vendor:
	@go mod vendor && go mod tidy
lint:
	@golangci-lint run

deploy:
	@./hack/rdeploy.sh

dev:
	@while true;do reflex -s -r ".*\.(html|go)$$" go run main.go;read -t1;done

setup-dev:
	@./hack/setup-local-dev.sh

fmt:
	@go fmt ./...

static:
	@go run ./main.go -gen ./html
