#!/usr/bin/env bash
set -x
set -e

git pull
[[ -x /tmp/chmouphotos ]] || time go build -o /tmp/chmouphoto main.go
sudo systemctl stop chmouphotos
sudo mv /tmp/chmouphotos /usr/local/bin/chmouphotos
sudo systemctl start chmouphotos
journalctl -u chmouphotos.service --lines=10 --no-pager
