#!/usr/bin/env bash
set -e

rpi=pi.lan
targetDir="~/GIT/chmouphoto"
[[ -n "$(git status --porcelain=v1)" ]] && {
    echo "You have local change(s), commit push them first"
    git --no-pager status 
    exit
}
git push
ssh ${rpi} "cd ${targetDir};./hack/redeploy.sh"

