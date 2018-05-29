#!/bin/bash

VERSION=18.03.1
DOCKER_FILE=docker-$VERSION-ce.tgz
DOCKER_SERVICE_FILE=/usr/lib/systemd/system/docker.service
DATA_ROOT=
function write_docker_service_file()
{
echo "
[Unit]
Description=dockerd service

[Service]
ExecStart=/usr/local/bin/dockerd $DATA_ROOT
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target" > $DOCKER_SERVICE_FILE
}

function download_docker()
{
	wget https://download.docker.com/linux/static/stable/x86_64/$DOCKER_FILE
	tar -xvzf $DOCKER_FILE
}

function install_docker()
{
	write_docker_service_file
	cp -rf docker/* /usr/local/bin/
}

function start_docker()
{
	systemctl enable docker
	systemctl start docker
	systemctl status docker
}

NUM_ARGS=$#

if [ $NUM_ARGS -gt 0 ]; then
	DATA_ROOT="--data-root $1"
	mkdir -p $1
fi

cd /tmp/
download_docker
install_docker
start_docker

cat $DOCKER_SERVICE_FILE
