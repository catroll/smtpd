import smtplib
from email.mime.text import MIMEText
from email.header import Header

# 服务器配置
smtp_server = "127.0.0.1"
smtp_port = 2525

# 认证信息
username = "user1"
password = "password123"

# 创建邮件
msg = MIMEText("这是一封测试邮件", "plain", "utf-8")
msg["Subject"] = Header("测试邮件主题", "utf-8")
msg["From"] = f"{username}@localhost"
msg["To"] = "recipient@localhost"

try:
    # 创建 SMTP 连接
    smtp = smtplib.SMTP(smtp_server, smtp_port)
    smtp.set_debuglevel(1)  # 开启调试模式
    
    # 登录（无需 STARTTLS）
    smtp.login(username, password)
    
    # 发送邮件
    smtp.sendmail(msg["From"], [msg["To"]], msg.as_string())
    print("邮件发送成功！")
    
except Exception as e:
    print(f"发送失败：{e}")
    
finally:
    try:
        smtp.quit()
    except:
        pass
