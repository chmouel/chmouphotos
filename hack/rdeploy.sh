#!/usr/bin/env bash
set -x
set -e

rpi=pi.lan
targetDir="~/GIT/chmouphoto"
ssh ${rpi} "cd ${targetDir};./hack/redeploy.sh"

