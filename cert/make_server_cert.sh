#!/bin/bash
set -e

. "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/params.sh"

SERVER_HOST=$1
CERT_NAME=${SERVER_HOST}
#shift
names=$*

if [[ -z "${CERT_NAME}" || -z "${SERVER_HOST}" ]]; then
  echo "usage: $0 SERVER_HOST [ALT_NAME...]"
  exit 1
fi

if [[ ! -e ${CA_NAME}.key ]]; then
  echo "No ca cert found!"
  exit 1
fi

# make server cert
openssl req -sha256 -nodes -newkey rsa:2048 -out ${CERT_NAME}.csr -keyout ${CERT_NAME}.key \
  -subj "${SUBJ}/CN=${SERVER_HOST}"

cat >ext.cfg <<-EOT
basicConstraints=critical,CA:TRUE
keyUsage = critical,digitalSignature,keyEncipherment,cRLSign,keyCertSign
extendedKeyUsage = critical,clientAuth,serverAuth
EOT

if [[ -n "$names" ]]; then
  for d in $names; do
    [[ -n "$s" ]] && s="$s,"

    if [[ "$d" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]]; then
      s="${s}IP:${d}"
    else
      s="${s}DNS:${d}"
    fi
  done
  echo "subjectAltName=$s" >>ext.cfg
fi

cat ext.cfg

openssl x509 -req -in ${CERT_NAME}.csr -CA ${CA_NAME}.pem -CAkey ${CA_NAME}.key -CAcreateserial \
  -out ${CERT_NAME}.pem -days 3650 -extfile ext.cfg

rm ext.cfg ${CERT_NAME}.csr

cat ${CERT_NAME}.pem ${CA_NAME}.pem > ${CERT_NAME}-chain.pem

p12name=truststore-srv.p12

[[ -e ${p12name} ]] && rm ${p12name}
openssl pkcs12 -export -nokeys -name ca -in ${CERT_NAME}-chain.pem -out ${p12name} -passout pass:${PASS}
