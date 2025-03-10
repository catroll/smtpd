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
go build -o smtpd.exe .
./smtpd.exe
perl swaks --to user@example.com --server localhost:1025
```
