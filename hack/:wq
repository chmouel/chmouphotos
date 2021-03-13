#!/bin/bash
set -x
set -e

git pull 
go build -o /tmp/chmouphoto main.go 
sudo systemctl stop chmouphoto 
sudo mv /tmp/chmouphoto /usr/local/bin/chmouphoto 
sudo systemctl start chmouphoto
