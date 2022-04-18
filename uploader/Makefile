RPI_HOST = kodi

all: redeploy


fmt:
	@go fmt `go list ./... | grep -v /vendor/`

redeploy:
	@sudo systemctl stop chmouphotos;go build -o /tmp/chmouphotos main.go && \
		sudo mv -fv /tmp/chmouphotos /usr/local/bin/chmouphotos && \
		sudo systemctl start chmouphotos && sudo systemctl status chmouphotos

cross-rpi:
	@env GOOS=linux GOARCH=arm GOARM=7 go build -o /tmp/chmouphotos-rpi

sync: cross-rpi
	ssh $(RPI_HOST) sudo systemctl stop chmouphotos && \
	scp /tmp/chmouphotos-rpi $(RPI_HOST):/usr/local/bin/chmouphotos && \
	ssh $(RPI_HOST) sudo systemctl stop chmouphotos && \
	ssh $(RPI_HOST) sudo systemctl status chmouphotos
