#!/usr/bin/env bash
set -x
set -e

rpi=pi.lan
targetDir="~/GIT/chmouphoto"

env GOOS=linux GOARCH=arm GOARM=5 go build -o /tmp/rpi-chmouphoto
ssh ${rpi} "cd ${targetDir};git pull "
scp /tmp/rpi-chmouphoto ${rpi}:/tmp/chmouphoto 
ssh ${rpi} "sudo systemctl stop chmouphoto && sudo mv /tmp/chmouphoto /usr/local/bin/chmouphoto && sudo systemctl start chmouphoto && journalctl -u chmouphoto.service --lines=10 --no-pager"

