#!/usr/bin/env bash
set -x
set -e
set -u

WEBDIR=/home/www/photos/

git pull
[[ -x /tmp/chmouphotos ]] || time go build -o /tmp/chmouphoto main.go
sudo systemctl stop chmouphotos
sudo mv /tmp/chmouphotos /usr/local/bin/chmouphotos

rsync --delete -avuz ./html ${WEBDIR}

sudo systemctl start chmouphotos
journalctl -u chmouphotos.service --lines=10 --no-pager
