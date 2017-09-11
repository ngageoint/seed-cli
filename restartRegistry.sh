#!/usr/bin/env bash
cd "$(dirname "$0")"
pwd
docker rm $(docker ps -a -q  --filter ancestor=registry:2) -f
docker run -d -p 5000:5000 --restart=always --name registry -v `pwd`/auth:/auth -e "REGISTRY_AUTH=htpasswd" -e "REGISTRY_AUTH_HTPASSWD_REALM=Registry Realm" -e REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd registry:2