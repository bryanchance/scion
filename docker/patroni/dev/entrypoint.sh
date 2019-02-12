#!/bin/bash

# Patroni entry point based on https://github.com/zalando/patroni

DOCKER_IP=$(hostname --ip-address)

export PATRONI_RESTAPI_CONNECT_ADDRESS="${DOCKER_IP}:8008"
export PATRONI_RESTAPI_LISTEN="0.0.0.0:8008"
export PATRONI_POSTGRESQL_CONNECT_ADDRESS="${DOCKER_IP}:5432"
export PATRONI_POSTGRESQL_LISTEN="0.0.0.0:5432"

exec python3 /patroni.py /patroni.yml

while true; do
    sleep 60
done
