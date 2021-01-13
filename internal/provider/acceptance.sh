#!/bin/bash

set -eu
set -x

cleanup_docker() {
	docker stop ns
	docker stop kdc || :
}
failed() {
	cleanup_docker
	exit 1
}

docker buildx build --target kdc --tag kdc internal/provider/testdata/
docker buildx build --target ns --tag ns internal/provider/testdata/
docker buildx build --target keytab --output type=local,dest=internal/provider/testdata/ internal/provider/testdata/

export DNS_UPDATE_SERVER=127.0.0.1
export DNS_UPDATE_PORT=53

# Run with no authentication

docker run -d --tmpfs /tmp --tmpfs /run \
	-v /sys/fs/cgroup:/sys/fs/cgroup:ro \
	-v /etc/localtime:/etc/localtime:ro \
	-v $PWD/internal/provider/testdata/named.conf.none:/etc/named.conf:ro \
	-p 127.0.0.1:53:53 \
	-p 127.0.0.1:53:53/udp \
	--rm --name ns --hostname ns.example.com ns || failed
GO111MODULE=on GOFLAGS=-mod=vendor make testacc TEST=./internal/provider || failed
cleanup_docker

# Run with TSIG authentication

docker run -d --tmpfs /tmp --tmpfs /run \
	-v /sys/fs/cgroup:/sys/fs/cgroup:ro \
	-v /etc/localtime:/etc/localtime:ro \
	-v $PWD/internal/provider/testdata/named.conf.tsig:/etc/named.conf:ro \
	-p 127.0.0.1:53:53 \
	-p 127.0.0.1:53:53/udp \
	--rm --name ns --hostname ns.example.com ns || failed
DNS_UPDATE_KEYNAME="tsig.example.com." DNS_UPDATE_KEYALGORITHM="hmac-sha256" DNS_UPDATE_KEYSECRET="UHeh4Iv/DVmPhi6LqCPDs6PixnyjLH4fjGESBjYnOyE=" GO111MODULE=on GOFLAGS=-mod=vendor make testacc TEST=./internal/provider || failed
cleanup_docker

# Run with Kerberos authentication

docker run -d --tmpfs /tmp --tmpfs /run \
	-v /sys/fs/cgroup:/sys/fs/cgroup:ro \
	-v /etc/localtime:/etc/localtime:ro \
	-p 127.0.0.1:88:88 \
	-p 127.0.0.1:88:88/udp \
	-p 127.0.0.1:464:464 \
	-p 127.0.0.1:464:464/udp \
	--rm --name kdc kdc || failed
docker run -d --tmpfs /tmp --tmpfs /run \
	-v /sys/fs/cgroup:/sys/fs/cgroup:ro \
	-v /etc/localtime:/etc/localtime:ro \
	-v $PWD/internal/provider/testdata/named.conf.kerberos:/etc/named.conf:ro \
	-p 127.0.0.1:53:53 \
	-p 127.0.0.1:53:53/udp \
	--rm --name ns --hostname ns.example.com ns || failed
# FIXME Run Kerberos tests here

cleanup_docker
