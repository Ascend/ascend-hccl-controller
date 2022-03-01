#!/usr/bin/env bash

set -o errexit

readonly caPath=${CA_PATH:-/etc/kubeedge/ca}
readonly subject=${SUBJECT:-/C=CN/ST=Sichuan/L=Chengdu/O=Huawei/OU=Ascend/CN=MindX}

genCA() {
    openssl ecparam -name secp384r1 -genkey -noout -out ${caPath}/rootCA.key
    openssl req -x509 -new -nodes -sha256 -days 3650 -subj ${subject} -key ${caPath}/rootCA.key -out ${caPath}/rootCA.crt
    chmod 400 ${caPath}/rootCA.*
}

ensureCA() {
    if [ ! -e ${caPath}/rootCA.key ] || [ ! -e ${caPath}/rootCA.crt ]; then
        genCA
    fi
}

ensureFolder() {
    if [ ! -d ${caPath} ]; then
        mkdir -p -m 700 ${caPath}
    fi
    if [ ! -d ${certPath} ]; then
        mkdir -p -m 700 ${certPath}
    fi
}

genCsr() {
    local name=$1
    openssl genrsa -out ${certPath}/${name}.key 4096
    openssl req -sha512 -new -subj ${subject} -key ${certPath}/${name}.key -out ${certPath}/${name}.csr
}

genCert() {
    local name=$1
    cat > ${certPath}/v3.ext <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
IP={{ HARBOR_IP }}
EOF

    openssl x509 -req -sha512 -days 3650 -extfile ${certPath}/v3.ext -CA ${caPath}/rootCA.crt -CAkey ${caPath}/rootCA.key \
    -CAcreateserial -in ${certPath}/${name}.csr -out ${certPath}/${name}.crt
    chmod 400 ${certPath}/${name}.* ${certPath}/v3.ext
    chmod 400 ${caPath}/rootCA.srl
}

dockerCert() {
    local docker_harbor=/etc/docker/certs.d/{{ HARBOR_IP }}:{{HARBOR_HTTPS_PORT}}
    mkdir -p -m 700 ${docker_harbor}
    [[ -e ${docker_harbor}/rootCA.crt ]] && rm ${docker_harbor}/rootCA.crt
    cp ${caPath}/rootCA.crt ${docker_harbor}
    chmod 700 /etc/docker/certs.d
    chmod 400 ${docker_harbor}/rootCA.crt
    mkdir -p -m 750 /var/log/harbor
}

genCertAndKey() {
    local name=$1
    certPath=$2
    ensureFolder
    ensureCA
    genCsr $name
    genCert $name
    dockerCert
}

$1 $2 $3
