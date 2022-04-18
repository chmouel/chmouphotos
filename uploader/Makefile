all: redeploy


fmt:
	@go fmt `go list ./... | grep -v /vendor/`

redeploy:
	@sudo systemctl stop chmouphotos;go build -o /tmp/chmouphotos main.go && sudo mv -fv /tmp/chmouphotos /usr/local/bin/chmouphotos && sudo systemctl start chmouphotos && sudo systemctl status chmouphotos
