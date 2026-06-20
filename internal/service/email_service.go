package service

import (
	"context"
	"fmt"
	"html"
	"net/smtp"
	"net/url"
	"strings"

	"art-backend/internal/config"
	"art-backend/internal/model"
)

type OrderEmailSender interface {
	SendOrderConfirmation(ctx context.Context, order model.Order) error
}

type SMTPEmailService struct {
	config config.Config
}

func NewSMTPEmailService(config config.Config) *SMTPEmailService {
	return &SMTPEmailService{config: config}
}

func (s *SMTPEmailService) SendOrderConfirmation(ctx context.Context, order model.Order) error {
	if !s.isConfigured() || order.CustomerEmail == nil || strings.TrimSpace(*order.CustomerEmail) == "" {
		return nil
	}

	to := strings.TrimSpace(*order.CustomerEmail)
	subject := "Order confirmation " + order.OrderNumber
	body := s.orderConfirmationHTML(order)
	message := strings.Join([]string{
		"From: " + s.fromHeader(),
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"",
		body,
	}, "\r\n")

	done := make(chan error, 1)
	go func() {
		addr := s.config.SMTPHost + ":" + s.config.SMTPPort
		auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
		done <- smtp.SendMail(addr, auth, s.config.SMTPFromEmail, []string{to}, []byte(message))
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (s *SMTPEmailService) isConfigured() bool {
	return strings.TrimSpace(s.config.SMTPHost) != "" &&
		strings.TrimSpace(s.config.SMTPPort) != "" &&
		strings.TrimSpace(s.config.SMTPUsername) != "" &&
		strings.TrimSpace(s.config.SMTPPassword) != "" &&
		strings.TrimSpace(s.config.SMTPFromEmail) != ""
}

func (s *SMTPEmailService) fromHeader() string {
	fromEmail := strings.TrimSpace(s.config.SMTPFromEmail)
	fromName := strings.TrimSpace(s.config.SMTPFromName)
	if fromName == "" {
		return fromEmail
	}
	return fmt.Sprintf("%s <%s>", fromName, fromEmail)
}

func (s *SMTPEmailService) orderConfirmationHTML(order model.Order) string {
	trackingURL := s.trackingURL(order.OrderNumber)
	customerName := "there"
	if order.CustomerName != nil && strings.TrimSpace(*order.CustomerName) != "" {
		customerName = strings.TrimSpace(*order.CustomerName)
	}

	var items strings.Builder
	for _, item := range order.Items {
		items.WriteString("<li>")
		items.WriteString(html.EscapeString(item.ProductTitle))
		if item.VariantSize != "" {
			items.WriteString(" - ")
			items.WriteString(html.EscapeString(item.VariantSize))
		}
		items.WriteString(fmt.Sprintf(" x%d - %s", item.Quantity, formatOrderMoney(item.Subtotal, order.Currency)))
		items.WriteString("</li>")
	}
	if len(order.Items) == 0 {
		items.WriteString("<li>Your artwork order</li>")
	}

	return fmt.Sprintf(`
		<p>Hi %s,</p>
		<p>Thank you for your order. Your payment has been received.</p>
		<p><strong>Order number:</strong> %s</p>
		<ul>%s</ul>
		<p>
			Subtotal: %s<br>
			Tax: %s<br>
			Shipping: %s<br>
			<strong>Total: %s</strong>
		</p>
		<p>You can track your order here: <a href="%s">%s</a></p>
	`, html.EscapeString(customerName), html.EscapeString(order.OrderNumber), items.String(), formatOrderMoney(order.Subtotal, order.Currency), formatOrderMoney(order.TaxAmount, order.Currency), formatOrderMoney(order.ShippingAmount, order.Currency), formatOrderMoney(order.TotalAmount, order.Currency), html.EscapeString(trackingURL), html.EscapeString(trackingURL))
}

func (s *SMTPEmailService) trackingURL(orderNumber string) string {
	baseURL := strings.TrimSpace(s.config.FrontendOrderTrackingURL)
	if baseURL == "" {
		baseURL = "http://localhost:5174/orders/track"
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	values := parsedURL.Query()
	values.Set("order_number", orderNumber)
	parsedURL.RawQuery = values.Encode()
	return parsedURL.String()
}

func formatOrderMoney(amount float64, currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}
	return fmt.Sprintf("%s %.2f", currency, amount)
}
