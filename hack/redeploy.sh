#!/usr/bin/env bash
set -x
set -e

git pull 
[[ -x /tmp/chmouphoto ]] || go build -o /tmp/chmouphoto main.go 
sudo systemctl stop chmouphoto 
sudo mv /tmp/chmouphoto /usr/local/bin/chmouphoto 
sudo systemctl start chmouphoto
journalctl -u chmouphoto.service --lines=10 --no-pager
