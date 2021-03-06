#!/usr/bin/env bash
# Chmouel Boudjnah <chmouel@chmouel.com>

set -ex

CONTAINER_NAME=mariadb
CONTAINER_IMAGE=mariadb
CONTAINER_HOST=127.0.0.1
CONTAINER_PORT=3306
MYSQL_ROOT_PASSWORD=chmouel
MYSQL_DATABASE=chmouphotos

REMOTE_SERVER=pi
REMOTE_PHOTOS=/home/www/photos/content/images
LOCAL_PHOTOS=$HOME/Pictures/photos/content/

rsync -avuz --delete ${REMOTE_SERVER}:${REMOTE_PHOTOS} ${LOCAL_PHOTOS}

sudo docker stop ${CONTAINER_NAME} || true && sudo docker rm ${CONTAINER_NAME}
sudo docker run -p ${CONTAINER_HOST}:${CONTAINER_PORT}:${CONTAINER_PORT} \
     --name ${CONTAINER_NAME} -d -e MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} ${CONTAINER_IMAGE}
sleep 10
mysqladmin -h${CONTAINER_HOST} -uroot -p${MYSQL_ROOT_PASSWORD} create ${MYSQL_DATABASE}
mysqldump --defaults-group-suffix=${REMOTE_SERVER} --add-drop-database ${MYSQL_DATABASE}|mysql ${MYSQL_DATABASE}
