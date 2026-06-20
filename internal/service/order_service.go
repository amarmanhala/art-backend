package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"art-backend/internal/config"
	"art-backend/internal/model"
)

var (
	ErrEmptyCart              = errors.New("empty cart")
	ErrInvalidStripeConfig    = errors.New("invalid stripe config")
	ErrInvalidStripeSignature = errors.New("invalid stripe signature")
	ErrUnsupportedStripeEvent = errors.New("unsupported stripe event")
	ErrOrderNotFound          = errors.New("order not found")
)

type OrderService struct {
	config     config.Config
	carts      CartStore
	orders     OrderStore
	emails     OrderEmailSender
	httpClient *http.Client
}

type CartStore interface {
	GetOrCreate(ctx context.Context, userID int64) (model.Cart, error)
	FindItems(ctx context.Context, cartID int64) ([]model.CartItem, error)
	Clear(ctx context.Context, cartID int64) error
}

type OrderStore interface {
	Create(ctx context.Context, request model.CreateOrderRequest) (model.Order, error)
	FindAllByUser(ctx context.Context, userID int64) ([]model.OrderSummary, error)
	FindByID(ctx context.Context, userID int64, id int64) (model.Order, error)
	FindByStripeSessionID(ctx context.Context, stripeSessionID string) (model.Order, error)
	FindByStripeSessionIDForUser(ctx context.Context, userID int64, stripeSessionID string) (model.Order, error)
	FindByOrderNumberAndCustomerEmail(ctx context.Context, orderNumber string, email string) (model.Order, error)
	UpdateStripeSession(ctx context.Context, orderID int64, stripeSessionID string) error
	MarkPaidByStripeSessionID(ctx context.Context, stripeSessionID string, paymentIntentID string, paymentStatus string, subtotal float64, taxAmount float64, shippingAmount float64, totalAmount float64, customerEmail *string, customerName *string, shippingName *string, shippingPhone *string, shippingLine1 *string, shippingLine2 *string, shippingCity *string, shippingState *string, shippingPostalCode *string, shippingCountry *string) (bool, error)
	MarkExpiredByStripeSessionID(ctx context.Context, stripeSessionID string) error
	MarkCheckoutFailed(ctx context.Context, orderID int64) error
}

func NewOrderService(config config.Config, carts CartStore, orders OrderStore) *OrderService {
	return &OrderService{
		config: config,
		carts:  carts,
		orders: orders,
		emails: NewSMTPEmailService(config),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *OrderService) CreateCheckoutSession(ctx context.Context, userID int64) (model.CheckoutSessionResponse, error) {
	if !s.isCheckoutConfigured() {
		return model.CheckoutSessionResponse{}, ErrInvalidStripeConfig
	}

	cart, err := s.carts.GetOrCreate(ctx, userID)
	if err != nil {
		return model.CheckoutSessionResponse{}, err
	}

	items, err := s.carts.FindItems(ctx, cart.ID)
	if err != nil {
		return model.CheckoutSessionResponse{}, err
	}
	if len(items) == 0 {
		return model.CheckoutSessionResponse{}, ErrEmptyCart
	}

	currency := strings.ToLower(strings.TrimSpace(items[0].Product.Currency))
	if currency == "" {
		currency = "usd"
	}
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item.Product.Currency)) != currency {
			return model.CheckoutSessionResponse{}, ErrInvalidStripeConfig
		}
	}

	orderNumber, err := generateOrderNumber()
	if err != nil {
		return model.CheckoutSessionResponse{}, err
	}

	orderRequest := model.CreateOrderRequest{
		UserID:      userID,
		OrderNumber: orderNumber,
		Currency:    strings.ToUpper(currency),
	}

	for _, item := range items {
		orderRequest.Subtotal += item.Subtotal
		orderRequest.Items = append(orderRequest.Items, model.CreateOrderItemRequest{
			ProductID:        item.Product.ID,
			ProductVariantID: item.Variant.ID,
			ProductTitle:     item.Product.Title,
			ProductSlug:      item.Product.Slug,
			VariantSize:      item.Variant.Size,
			UnitPrice:        item.Variant.Price,
			Quantity:         item.Quantity,
			Subtotal:         item.Subtotal,
			ImageURL:         item.Product.ImageURL,
			ThumbnailURL:     item.Product.ThumbnailURL,
		})
	}
	orderRequest.TotalAmount = orderRequest.Subtotal

	order, err := s.orders.Create(ctx, orderRequest)
	if err != nil {
		return model.CheckoutSessionResponse{}, err
	}

	session, err := s.createStripeCheckoutSession(ctx, order, items)
	if err != nil {
		_ = s.orders.MarkCheckoutFailed(ctx, order.ID)
		return model.CheckoutSessionResponse{}, err
	}

	if err := s.orders.UpdateStripeSession(ctx, order.ID, session.ID); err != nil {
		return model.CheckoutSessionResponse{}, err
	}

	expiresAt := time.Now().UTC().Add(30 * time.Minute)
	return model.CheckoutSessionResponse{
		CheckoutURL: session.URL,
		SessionID:   session.ID,
		OrderID:     order.ID,
		OrderNumber: order.OrderNumber,
		ExpiresAt:   expiresAt,
	}, nil
}

func (s *OrderService) GetAll(ctx context.Context, userID int64) ([]model.OrderSummary, error) {
	return s.orders.FindAllByUser(ctx, userID)
}

func (s *OrderService) GetByID(ctx context.Context, userID int64, id int64) (model.Order, error) {
	if id <= 0 {
		return model.Order{}, ErrOrderNotFound
	}
	order, err := s.orders.FindByID(ctx, userID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Order{}, ErrOrderNotFound
	}
	return order, err
}

func (s *OrderService) GetByStripeSessionID(ctx context.Context, userID int64, stripeSessionID string) (model.Order, error) {
	stripeSessionID = strings.TrimSpace(stripeSessionID)
	if userID <= 0 || stripeSessionID == "" {
		return model.Order{}, ErrOrderNotFound
	}

	order, err := s.orders.FindByStripeSessionIDForUser(ctx, userID, stripeSessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Order{}, ErrOrderNotFound
	}
	return order, err
}

func (s *OrderService) TrackOrder(ctx context.Context, request model.TrackOrderRequest) (model.Order, error) {
	orderNumber := strings.TrimSpace(request.OrderNumber)
	email := strings.ToLower(strings.TrimSpace(request.Email))
	if orderNumber == "" || email == "" || !strings.Contains(email, "@") {
		return model.Order{}, ErrOrderNotFound
	}

	order, err := s.orders.FindByOrderNumberAndCustomerEmail(ctx, orderNumber, email)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Order{}, ErrOrderNotFound
	}
	return order, err
}

func (s *OrderService) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.isWebhookConfigured() {
		return ErrInvalidStripeConfig
	}

	event, err := verifyStripeWebhookEvent(payload, signature, s.config.StripeWebhookSecret)
	if err != nil {
		return err
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripeCheckoutSession
		if err := json.Unmarshal(event.Data.Object, &session); err != nil {
			return err
		}

		if strings.ToLower(session.PaymentStatus) != "paid" {
			return nil
		}

		customerEmail := nullStringFrom(session.CustomerDetails.Email, session.CustomerEmail)
		customerName := nullStringFrom(session.CustomerDetails.Name, "")
		shippingName := nullStringFrom(session.ShippingDetails.Name, "")
		shippingPhone := nullStringFrom(session.ShippingDetails.Phone, "")
		shippingLine1 := nullStringFrom(session.ShippingDetails.Address.Line1, "")
		shippingLine2 := nullStringFrom(session.ShippingDetails.Address.Line2, "")
		shippingCity := nullStringFrom(session.ShippingDetails.Address.City, "")
		shippingState := nullStringFrom(session.ShippingDetails.Address.State, "")
		shippingPostalCode := nullStringFrom(session.ShippingDetails.Address.PostalCode, "")
		shippingCountry := nullStringFrom(session.ShippingDetails.Address.Country, "")
		amounts := stripeOrderAmountsFromSession(session)

		updated, err := s.orders.MarkPaidByStripeSessionID(
			ctx,
			session.ID,
			session.PaymentIntent,
			session.PaymentStatus,
			amounts.Subtotal,
			amounts.TaxAmount,
			amounts.ShippingAmount,
			amounts.TotalAmount,
			customerEmail,
			customerName,
			shippingName,
			shippingPhone,
			shippingLine1,
			shippingLine2,
			shippingCity,
			shippingState,
			shippingPostalCode,
			shippingCountry,
		)
		if err != nil || !updated {
			return err
		}

		order, err := s.orders.FindByStripeSessionID(ctx, session.ID)
		if err != nil {
			return err
		}

		s.sendOrderConfirmation(ctx, order)

		return s.carts.Clear(ctx, order.UserID)
	case "checkout.session.expired":
		var session stripeCheckoutSession
		if err := json.Unmarshal(event.Data.Object, &session); err != nil {
			return err
		}
		return s.orders.MarkExpiredByStripeSessionID(ctx, session.ID)
	default:
		return nil
	}
}

func (s *OrderService) sendOrderConfirmation(ctx context.Context, order model.Order) {
	if s.emails == nil || order.CustomerEmail == nil || strings.TrimSpace(*order.CustomerEmail) == "" {
		return
	}
	if err := s.emails.SendOrderConfirmation(ctx, order); err != nil {
		log.Printf("could not send order confirmation email for order %s: %v", order.OrderNumber, err)
	}
}

func (s *OrderService) isCheckoutConfigured() bool {
	return s.config.StripeSecretKey != "" &&
		s.config.StripeSuccessURL != "" &&
		s.config.StripeCancelURL != ""
}

func (s *OrderService) isWebhookConfigured() bool {
	return s.config.StripeWebhookSecret != ""
}

func (s *OrderService) createStripeCheckoutSession(ctx context.Context, order model.Order, items []model.CartItem) (stripeCheckoutSessionResponse, error) {
	values := url.Values{}
	values.Set("mode", "payment")
	values.Set("success_url", s.config.StripeSuccessURL)
	values.Set("cancel_url", s.config.StripeCancelURL)
	values.Set("client_reference_id", order.OrderNumber)
	values.Set("metadata[order_id]", strconv.FormatInt(order.ID, 10))
	values.Set("metadata[order_number]", order.OrderNumber)
	values.Set("metadata[user_id]", strconv.FormatInt(order.UserID, 10))
	values.Set("automatic_tax[enabled]", "true")
	for _, country := range s.allowedShippingCountries() {
		values.Add("shipping_address_collection[allowed_countries][]", country)
	}

	for i, item := range items {
		prefix := fmt.Sprintf("line_items[%d]", i)
		unitAmount := int64(math.Round(item.Variant.Price * 100))
		values.Set(prefix+"[quantity]", strconv.Itoa(item.Quantity))
		values.Set(prefix+"[price_data][currency]", strings.ToLower(order.Currency))
		values.Set(prefix+"[price_data][unit_amount]", strconv.FormatInt(unitAmount, 10))
		values.Set(prefix+"[price_data][tax_behavior]", "exclusive")
		values.Set(prefix+"[price_data][product_data][name]", sessionProductName(item))
		values.Set(prefix+"[price_data][product_data][tax_code]", s.productTaxCode())
		if item.Product.ImageURL != "" {
			values.Add(prefix+"[price_data][product_data][images][]", item.Product.ImageURL)
		}
		if item.Variant.Size != "" {
			values.Set(prefix+"[price_data][product_data][description]", "Size: "+item.Variant.Size)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.stripe.com/v1/checkout/sessions", strings.NewReader(values.Encode()))
	if err != nil {
		return stripeCheckoutSessionResponse{}, err
	}
	req.SetBasicAuth(s.config.StripeSecretKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return stripeCheckoutSessionResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return stripeCheckoutSessionResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return stripeCheckoutSessionResponse{}, fmt.Errorf("stripe checkout session error: %s", strings.TrimSpace(string(body)))
	}

	var session stripeCheckoutSessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return stripeCheckoutSessionResponse{}, err
	}
	if session.ID == "" || session.URL == "" {
		return stripeCheckoutSessionResponse{}, fmt.Errorf("stripe checkout session response missing id or url")
	}

	return session, nil
}

func (s *OrderService) productTaxCode() string {
	taxCode := strings.TrimSpace(s.config.StripeProductTaxCode)
	if taxCode == "" {
		return "txcd_99999999"
	}
	return taxCode
}

func (s *OrderService) allowedShippingCountries() []string {
	raw := strings.TrimSpace(s.config.StripeAllowedShippingCountries)
	if raw == "" {
		return []string{"US", "CA"}
	}

	parts := strings.Split(raw, ",")
	countries := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ToUpper(strings.TrimSpace(part))
		if part != "" {
			countries = append(countries, part)
		}
	}
	if len(countries) == 0 {
		return []string{"US", "CA"}
	}

	return countries
}

func generateOrderNumber() (string, error) {
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return fmt.Sprintf("ORD-%s-%s", time.Now().UTC().Format("20060102"), strings.ToUpper(hex.EncodeToString(randomBytes))), nil
}

func sessionProductName(item model.CartItem) string {
	if item.Variant.Size == "" {
		return item.Product.Title
	}

	return item.Product.Title + " - " + item.Variant.Size
}

func verifyStripeWebhookEvent(payload []byte, signature string, secret string) (stripeWebhookEvent, error) {
	if secret == "" {
		return stripeWebhookEvent{}, ErrInvalidStripeConfig
	}

	timestamp, signatures, err := parseStripeSignature(signature)
	if err != nil {
		return stripeWebhookEvent{}, ErrInvalidStripeSignature
	}

	expected := computeStripeSignature(timestamp, payload, secret)
	valid := false
	for _, candidate := range signatures {
		if hmac.Equal([]byte(candidate), []byte(expected)) {
			valid = true
			break
		}
	}
	if !valid {
		return stripeWebhookEvent{}, ErrInvalidStripeSignature
	}

	var event stripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return stripeWebhookEvent{}, err
	}
	return event, nil
}

func parseStripeSignature(header string) (string, []string, error) {
	parts := strings.Split(header, ",")
	var timestamp string
	signatures := make([]string, 0)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		}
		if strings.HasPrefix(part, "v1=") {
			signatures = append(signatures, strings.TrimPrefix(part, "v1="))
		}
	}
	if timestamp == "" || len(signatures) == 0 {
		return "", nil, errors.New("invalid stripe signature header")
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", nil, err
	}
	delta := time.Since(time.Unix(ts, 0))
	if delta > 5*time.Minute || delta < -5*time.Minute {
		return "", nil, errors.New("stripe signature timestamp outside tolerance")
	}

	return timestamp, signatures, nil
}

func computeStripeSignature(timestamp string, payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func stripeAmountToDecimal(amount int64) float64 {
	return float64(amount) / 100
}

func stripeOrderAmountsFromSession(session stripeCheckoutSession) stripeOrderAmounts {
	return stripeOrderAmounts{
		Subtotal:       stripeAmountToDecimal(session.AmountSubtotal),
		TaxAmount:      stripeAmountToDecimal(session.TotalDetails.AmountTax),
		ShippingAmount: stripeAmountToDecimal(session.TotalDetails.AmountShipping),
		TotalAmount:    stripeAmountToDecimal(session.AmountTotal),
	}
}

type stripeOrderAmounts struct {
	Subtotal       float64
	TaxAmount      float64
	ShippingAmount float64
	TotalAmount    float64
}

func nullStringFrom(values ...string) *string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			v := value
			return &v
		}
	}
	return nil
}

type stripeCheckoutSessionResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type stripeWebhookEvent struct {
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

type stripeCheckoutSession struct {
	ID             string            `json:"id"`
	AmountSubtotal int64             `json:"amount_subtotal"`
	AmountTotal    int64             `json:"amount_total"`
	PaymentIntent  string            `json:"payment_intent"`
	PaymentStatus  string            `json:"payment_status"`
	CustomerEmail  string            `json:"customer_email"`
	Metadata       map[string]string `json:"metadata"`
	AutomaticTax   struct {
		Status string `json:"status"`
	} `json:"automatic_tax"`
	TotalDetails struct {
		AmountDiscount int64 `json:"amount_discount"`
		AmountShipping int64 `json:"amount_shipping"`
		AmountTax      int64 `json:"amount_tax"`
	} `json:"total_details"`
	CustomerDetails struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"customer_details"`
	ShippingDetails struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Address struct {
			Line1      string `json:"line1"`
			Line2      string `json:"line2"`
			City       string `json:"city"`
			State      string `json:"state"`
			PostalCode string `json:"postal_code"`
			Country    string `json:"country"`
		} `json:"address"`
	} `json:"shipping_details"`
}
