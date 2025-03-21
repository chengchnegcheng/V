package notification

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
}

// NewEmailNotifier 创建新的邮件通知器
func NewEmailNotifier(host string, port int, username, password, from, fromName string) *EmailNotifier {
	return &EmailNotifier{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		fromName: fromName,
	}
}

// Send 发送通知
func (n *EmailNotifier) Send(notification *Notification) error {
	// 如果没有收件人，则跳过
	if len(notification.To) == 0 {
		return nil
	}

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", n.fromName, n.from)
	headers["To"] = strings.Join(notification.To, ", ")
	headers["Subject"] = notification.Subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建邮件内容
	var message string
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + notification.Body

	// 设置SMTP连接信息
	auth := smtp.PlainAuth("", n.username, n.password, n.host)
	addr := fmt.Sprintf("%s:%d", n.host, n.port)

	// 发送邮件
	err := smtp.SendMail(addr, auth, n.from, notification.To, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
