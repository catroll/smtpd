# smtpd

## 功能

- SMTP 服务器
- StartTLS / SMTPS
- 认证
- 访问控制
- 额度控制
- 从配置中心获取配置
- 日志
- 监控与告警（性能统计）

## 开发与调试

```sh
go build -mod=readonly -o smtpd.exe .
./smtpd.exe
perl swaks --to user@example.com --server localhost:1025
```

生成自签名证书：

```sh
openssl genpkey -algorithm RSA -out server.key -pkeyopt rsa_keygen_bits:2048
openssl req -new -key server.key -out server.csr -subj "/C=US/ST=California/L=San Francisco/O=My Company/OU=IT/CN=catroll.com"
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt
```
