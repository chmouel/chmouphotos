#!/usr/bin/env bash
set -x
set -e

rpi=pi.lan
targetDir="~/GIT/chmouphoto"

env GOOS=linux GOARCH=arm GOARM=5 go build -o /tmp/rpi-chmouphoto
scp /tmp/rpi-chmouphoto ${rpi}:/tmp/chmouphoto 
ssh ${rpi} "cd ${targetDir};./hack/redeploy.sh"

