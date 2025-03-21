package notification

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"v/logger"
	"v/settings"
)

// Notification represents a notification
type Notification struct {
	To      []string
	Subject string
	Body    string
	Type    string
}

// Notifier defines the interface for notification services
type Notifier interface {
	Send(notification *Notification) error
}

// Manager represents a notification manager
type Manager struct {
	log      *logger.Logger
	settings *settings.Manager
}

// New creates a new notification manager
func New(log *logger.Logger, settings *settings.Manager) Notifier {
	return &Manager{
		log:      log,
		settings: settings,
	}
}

// Send sends a notification
func (m *Manager) Send(notification *Notification) error {
	// Get notification settings
	s := m.settings.Get()
	if !s.Notification.EnableEmail {
		return fmt.Errorf("email notifications are disabled")
	}

	// Validate SMTP settings
	if s.Notification.SMTPHost == "" || s.Notification.SMTPPort == 0 {
		return fmt.Errorf("SMTP settings are not configured")
	}

	// Send email
	if err := m.sendEmail(notification); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	m.log.Info("Notification sent", logger.Fields{
		"type":      notification.Type,
		"to":        notification.To,
		"subject":   notification.Subject,
		"timestamp": time.Now(),
	})

	return nil
}

// sendEmail sends an email notification
func (m *Manager) sendEmail(notification *Notification) error {
	s := m.settings.Get()

	// Prepare email
	from := fmt.Sprintf("%s <%s>", s.Notification.FromName, s.Notification.FromEmail)
	to := strings.Join(notification.To, ", ")
	subject := notification.Subject
	body := notification.Body

	// Create message
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", from, to, subject, body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.Notification.SMTPHost, s.Notification.SMTPPort)
	auth := smtp.PlainAuth("",
		s.Notification.SMTPUser,
		s.Notification.SMTPPassword,
		s.Notification.SMTPHost)

	// Send email
	if err := smtp.SendMail(addr, auth, s.Notification.FromEmail, notification.To, []byte(message)); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// SendTrafficWarning sends a traffic warning notification
func (m *Manager) SendTrafficWarning(userID int64, username string, usage, limit int64) error {
	s := m.settings.Get()
	warningPercent := float64(s.Traffic.WarningPercent) / 100
	usagePercent := float64(usage) / float64(limit)

	if usagePercent >= warningPercent {
		subject := "Traffic Usage Warning"
		body := fmt.Sprintf(`
			<p>Dear %s,</p>
			<p>Your traffic usage has reached %.1f%% of your limit.</p>
			<p>Current usage: %.2f GB</p>
			<p>Traffic limit: %.2f GB</p>
			<p>Please consider upgrading your plan or reducing your usage.</p>
			<p>Best regards,<br>%s</p>
		`, username, usagePercent*100, float64(usage)/1024/1024/1024, float64(limit)/1024/1024/1024, s.Site.Name)

		notification := &Notification{
			To:      []string{username},
			Subject: subject,
			Body:    body,
			Type:    "traffic_warning",
		}

		return m.Send(notification)
	}

	return nil
}

// SendExpirationWarning sends an account expiration warning notification
func (m *Manager) SendExpirationWarning(userID int64, username string, expireAt time.Time) error {
	daysLeft := int(time.Until(expireAt).Hours() / 24)
	if daysLeft <= 7 {
		subject := "Account Expiration Warning"
		body := fmt.Sprintf(`
			<p>Dear %s,</p>
			<p>Your account will expire in %d days.</p>
			<p>Expiration date: %s</p>
			<p>Please renew your account to continue using our services.</p>
			<p>Best regards,<br>%s</p>
		`, username, daysLeft, expireAt.Format("2006-01-02"), m.settings.Get().Site.Name)

		notification := &Notification{
			To:      []string{username},
			Subject: subject,
			Body:    body,
			Type:    "expiration_warning",
		}

		return m.Send(notification)
	}

	return nil
}

// SendCertificateExpirationWarning sends a certificate expiration warning notification
func (m *Manager) SendCertificateExpirationWarning(domain string, expireAt time.Time) error {
	daysLeft := int(time.Until(expireAt).Hours() / 24)
	if daysLeft <= 7 {
		subject := "SSL Certificate Expiration Warning"
		body := fmt.Sprintf(`
			<p>Dear Administrator,</p>
			<p>The SSL certificate for domain %s will expire in %d days.</p>
			<p>Expiration date: %s</p>
			<p>Please renew the certificate to maintain secure connections.</p>
			<p>Best regards,<br>%s</p>
		`, domain, daysLeft, expireAt.Format("2006-01-02"), m.settings.Get().Site.Name)

		notification := &Notification{
			To:      []string{m.settings.Get().SSL.Email},
			Subject: subject,
			Body:    body,
			Type:    "certificate_warning",
		}

		return m.Send(notification)
	}

	return nil
}

// SendBackupNotification sends a backup completion notification
func (m *Manager) SendBackupNotification(success bool, path string, size int64) error {
	subject := "Backup Completion"
	status := "successful"
	if !success {
		status = "failed"
	}

	body := fmt.Sprintf(`
		<p>Dear Administrator,</p>
		<p>The system backup has completed %s.</p>
		<p>Backup path: %s</p>
		<p>Backup size: %.2f GB</p>
		<p>Timestamp: %s</p>
		<p>Best regards,<br>%s</p>
	`, status, path, float64(size)/1024/1024/1024, time.Now().Format("2006-01-02 15:04:05"), m.settings.Get().Site.Name)

	notification := &Notification{
		To:      []string{m.settings.Get().SSL.Email},
		Subject: subject,
		Body:    body,
		Type:    "backup",
	}

	return m.Send(notification)
}
