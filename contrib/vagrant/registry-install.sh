#!/bin/bash

set -e

docker run -d -p 5000:5000 \
   -e REGISTRY_AUTH=token \
   -e REGISTRY_AUTH_TOKEN_REALM=https://hub.lstbknd.net/registry/auth \
   -e REGISTRY_AUTH_TOKEN_SERVICE=hub.lstbknd.net \
   -e REGISTRY_AUTH_TOKEN_ISSUER=api.lstbknd.net \
   -e REGISTRY_AUTH_TOKEN_ROOTCERTBUNDLE=/ssl/lstbknd.net.pem \
   -e REGISTRY_HTTP_TLS_CERTIFICATE=/ssl/lstbknd.net.pem \
   -e REGISTRY_HTTP_TLS_KEY=/ssl/lstbknd.net.key \
   -e REGISTRY_HTTP_SECRET=qwerty \
   -v /etc/lastbackend/registry:/var/lib/registry \
   -v /etc/lastbackend/ssl:/ssl \
   --restart=always \
   --name hub registry:2