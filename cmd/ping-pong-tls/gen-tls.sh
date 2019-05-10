DNS="localhost"
PASSWORD='123456'
DAYS=3650
SUBJ="/C=CN/ST=GD/L=SZ/O=Self Sign Corporation/OU=Self Sign Software Group/CN=Tester CA/emailAddress=self@localhost"

### 创建CA
openssl genrsa -aes256 -passout "pass:${PASSWORD}" -out ca-key.pem 4096
openssl req -new -x509 -days ${DAYS} -key ca-key.pem -passin "pass:${PASSWORD}" -sha256 -out ca.pem -subj "${SUBJ}"

### 创建 server key
openssl genrsa -out server-key.pem 4096
openssl req -subj "${SUBJ}" -sha256 -new -key server-key.pem -passin "pass:${PASSWORD}" -out server.csr

echo subjectAltName = DNS:$DNS,IP:127.0.0.1 > extfile-server.cnf
echo extendedKeyUsage = serverAuth >> extfile-server.cnf

#### 使用 server key 签名生成 server cert
openssl x509 -req -days ${DAYS} -sha256 -in server.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out server-cert.pem -extfile extfile-server.cnf -passin "pass:${PASSWORD}"

### 创建客户端CA
openssl genrsa -out key.pem 4096
openssl req -subj "${SUBJ}" -new -key key.pem -out client.csr -passin "pass:${PASSWORD}"
echo extendedKeyUsage = clientAuth > extfile-client.cnf

#### 使用 client key 签名生成 client cert
openssl x509 -req -days ${DAYS} -sha256 -in client.csr -CA ca.pem -CAkey ca-key.pem -CAcreateserial -out cert.pem -extfile extfile-client.cnf -passin "pass:${PASSWORD}"

### 修改权限
chmod 0400 ca.pem key.pem cert.pem
chmod 0444 ca.pem server-key.pem server-cert.pem

### 各归其位
rm -rf ca
mkdir ca
cp ca.pem ca-key.pem ca/

rm -rf server
mkdir -p server
cp ca.pem server-cert.pem server-key.pem ./server

rm -rf client
mkdir -p client
cp ca.pem cert.pem key.pem ./client

### 清理工作
rm -f client.csr server.csr extfile-server.cnf extfile-client.cnf
rm -f ca.srl
rm -f *.pem
