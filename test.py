import smtplib
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.mime.application import MIMEApplication

sender_email = "your_email@example.com"
sender_password = "your_email_password"
receiver_email = "recipient_email@example.com"
subject = "Test Email with Attachment"
body = "This is a test email sent with an attachment using Python."

message = MIMEMultipart()
message["From"] = sender_email
message["To"] = receiver_email
message["Subject"] = subject

message.attach(MIMEText(body, "plain"))

attachment_path = "./go.mod"
with open(attachment_path, "rb") as file:
    attachment = MIMEApplication(file.read(), _subtype="pdf")
    attachment.add_header(
        "Content-Disposition", "attachment", filename="attachment.pdf"
    )
    message.attach(attachment)

smtp_server = "localhost"
smtp_port = 2525

try:
    server = smtplib.SMTP(smtp_server, smtp_port)

    # server.starttls()
    # server.login(sender_email, sender_password)

    text = message.as_string()
    server.sendmail(sender_email, receiver_email, text)
    print("Email sent successfully!")
except Exception as e:
    print(f"Error: {e}")
finally:
    server.quit()
