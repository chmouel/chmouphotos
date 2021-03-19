#!/usr/bin/env bash
set -x
set -e

git pull 
[[ -x /tmp/chmouphoto ]] || /usr/local/go/bin/go build -o /tmp/chmouphoto main.go 
sudo systemctl stop chmouphotos
sudo mv /tmp/chmouphoto /usr/local/bin/chmouphoto 
sudo systemctl start chmouphotos
journalctl -u chmouphotos.service --lines=10 --no-pager
