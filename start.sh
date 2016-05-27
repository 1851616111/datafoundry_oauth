#!/bin/bash
function env::export() {
    env=$(eval echo \$$1)
    if [ "X$env" = "X" ];then

        if [ $1 = GITHUB_REDIRECT_URL ]; then
            export GITHUB_REDIRECT_URL=http://datafoundry-oauth.app.dataos.io/v1/repos/github-redirect
        fi
        if [ $1 = GITHUB_CLIENT_ID ]; then
            export GITHUB_CLIENT_ID=2369ed831a59847924b4
        fi
        if [ $1 = GITHUB_CLIENT_SECRET ]; then
            export GITHUB_CLIENT_SECRET=510bb29970fcd684d0e7136a5947f92710332c98
        fi

        if [ $1 = DATAFOUNDRY_HOST_ADDR ]; then
            export DATAFOUNDRY_HOST_ADDR=https://dev.dataos.io:8443
        fi

        if [ $1 = ETCD_HTTP_ADDR ]; then
            export ETCD_HTTP_ADDR=http://etcdsystem.servicebroker.dataos.io
        fi
        if [ $1 = ETCD_HTTP_PORT ]; then
            export ETCD_HTTP_PORT=2379
        fi
        if [ $1 = ETCD_USER ]; then
            export ETCD_USER=asiainfoLDP
        fi
        if [ $1 = ETCD_PASSWORD ]; then
            export ETCD_PASSWORD=6ED9BA74-75FD-4D1B-8916-842CB936AC1A
        fi
    fi
}

function Env::Exports() {
    for param in $*; do
        env::export $param
    done
}

Env::Exports GITHUB_REDIRECT_URL GITHUB_CLIENT_ID GITHUB_CLIENT_SECRET
Env::Exports DATAFOUNDRY_HOST_ADDR
Env::Exports ETCD_HTTP_ADDR ETCD_HTTP_PORT ETCD_USER ETCD_PASSWORD

./datafoundry_oauth2