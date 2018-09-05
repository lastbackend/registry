#!/bin/bash

docker run -d -p 5000:5000 \
-e REGISTRY_AUTH=token \
-e REGISTRY_AUTH_TOKEN_REALM=https://hub.lstbknd.net/registry/auth \
-e REGISTRY_AUTH_TOKEN_SERVICE=hub.lstbknd.net \
-e REGISTRY_AUTH_TOKEN_ISSUER=registry.lstbknd.net \
-e REGISTRY_AUTH_TOKEN_ROOTCERTBUNDLE=/ssl/hub.lstbknd.net.bundle.pem \
-e REGISTRY_HTTP_TLS_CERTIFICATE=/ssl/hub.lstbknd.net.cert.pem \
-e REGISTRY_HTTP_TLS_KEY=/ssl/hub.lstbknd.net.key.pem \
-v /opt/registry:/var/lib/registry \
-v /opt/ssl:/ssl \
--restart=always \
--name hub registry:2