lint:
	@golangci-lint run 

rdeploy:
	@./hack/rdeploy.sh

dev:
	@while true;do reflex -s -r ".*\.(html|go)$$" go run main.go;read -t1;done	

setup-dev:
	@./hack/setup-local-dev.sh
