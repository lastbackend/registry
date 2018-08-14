#!/usr/bin/env bash

set -e

# define location of openssl binary manually since running this
# script under Vagrant fails on some systems without it
OPENSSL=$(which openssl)


if [ -z "$VAR" ]; then
    HOSTNAME="localhost"
fi

echo "===> Creating directories for both the server and client certificate sets"

OUTDIR=./ssl

mkdir -p ${OUTDIR}

echo "===> Create and sign a CA key and certificate and copy the CA certificate into ${OUTDIR}\n"

# establish cluster CA and self-sign a cert
$OPENSSL genrsa -out ${OUTDIR}/ca-key.pem 2048

$OPENSSL req -x509 -new -nodes -key ${OUTDIR}/ca-key.pem -days 10000 -out ${OUTDIR}/ca.pem -subj '/CN=lb-ca'

echo "===> Configuration file for the client ${OUTDIR}/client.cnf"

bash -c "cat <<EOT > ${OUTDIR}/client.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
EOT"


echo "===> Configuration file for the server ${OUTDIR}/server.cnf"

bash -c "cat <<EOT > ${OUTDIR}/server.cnf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[v3_req]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
IP.1  = 127.0.0.1
EOT"


echo "\n"
echo "===> Create and sign a certificate for the client"

$OPENSSL genrsa -out ${OUTDIR}/client-key.pem 2048

$OPENSSL req -new -key ${OUTDIR}/client-key.pem -out ${OUTDIR}/client.csr \
    -subj '/CN=localhost' -config ${OUTDIR}/client.cnf

$OPENSSL x509 -req -in ${OUTDIR}/client.csr -CA ${OUTDIR}/ca.pem \
  -CAkey ${OUTDIR}/ca-key.pem -CAcreateserial \
  -out ${OUTDIR}/client.pem -days 365 -extensions v3_req \
  -extfile ${OUTDIR}/client.cnf


echo "\n"
echo "===> Create and sign a certificate for the server"

openssl genrsa -out ${OUTDIR}/server-key.pem 2048

openssl req -new -key ${OUTDIR}/server-key.pem \
  -out ${OUTDIR}/server.csr \
  -subj '/CN=localhost' -config ${OUTDIR}/server.cnf

openssl x509 -req -in ${OUTDIR}/server.csr -CA ${OUTDIR}/ca.pem \
  -CAkey ${OUTDIR}/ca-key.pem -CAcreateserial \
  -out ${OUTDIR}/server.pem -days 365 -extensions v3_req \
  -extfile ${OUTDIR}/server.cnf
echo "\n"