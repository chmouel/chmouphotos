#!/usr/bin/env bash
set -e

rpi=kodi
targetDir="./GIT/chmouphotos"
[[ -n "$(git status --porcelain=v1)" ]] && {
    echo "You have local change(s), commit push them first"
    git --no-pager status
    exit
}
git push

env GOOS=linux GOARCH=arm GOARM=7 go build -o /tmp/rpi-chmouphotos
scp /tmp/rpi-chmouphotos ${rpi}:/tmp/chmouphotos
ssh ${rpi} "cd ${targetDir};./hack/redeploy.sh"
