package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"art-backend/internal/config"
	"art-backend/internal/model"
)

func TestVerifyStripeWebhookEvent(t *testing.T) {
	payload := map[string]any{
		"id":   "evt_test",
		"type": "checkout.session.completed",
		"data": map[string]any{
			"object": map[string]any{
				"id":               "cs_test_123",
				"payment_intent":   "pi_test_123",
				"payment_status":   "paid",
				"customer_email":   "buyer@example.com",
				"customer_details": map[string]any{"email": "buyer@example.com", "name": "Buyer"},
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	secret := "whsec_test_secret"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := computeStripeSignature(timestamp, body, secret)
	header := "t=" + timestamp + ",v1=" + signature

	event, err := verifyStripeWebhookEvent(body, header, secret)
	if err != nil {
		t.Fatalf("verifyStripeWebhookEvent returned error: %v", err)
	}
	if event.Type != "checkout.session.completed" {
		t.Fatalf("unexpected event type: %s", event.Type)
	}

	if _, err := verifyStripeWebhookEvent(body, header, "wrong_secret"); err == nil {
		t.Fatal("expected signature verification to fail with wrong secret")
	}
}

func TestGenerateOrderNumber(t *testing.T) {
	orderNumber, err := generateOrderNumber()
	if err != nil {
		t.Fatalf("generateOrderNumber returned error: %v", err)
	}

	if !strings.HasPrefix(orderNumber, "ORD-") {
		t.Fatalf("expected ORD- prefix, got %q", orderNumber)
	}
}

func TestCreateStripeCheckoutSessionIncludesAutomaticTax(t *testing.T) {
	var form url.Values
	orderService := NewOrderService(config.Config{
		StripeSecretKey:                "sk_test_secret",
		StripeSuccessURL:               "https://example.com/success",
		StripeCancelURL:                "https://example.com/cancel",
		StripeAllowedShippingCountries: "US,CA",
		StripeProductTaxCode:           "txcd_test_art",
	}, nil, nil)
	orderService.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			form, err = url.ParseQuery(string(body))
			if err != nil {
				return nil, err
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"id":"cs_test_tax","url":"https://checkout.stripe.test/session"}`)),
			}, nil
		}),
	}

	_, err := orderService.createStripeCheckoutSession(context.Background(), model.Order{
		ID:          1,
		OrderNumber: "ORD-TEST",
		UserID:      10,
		Currency:    "USD",
	}, []model.CartItem{
		{
			Product: model.Product{
				ID:       1,
				Title:    "Canvas",
				ImageURL: "https://example.com/canvas.jpg",
			},
			Variant: model.ProductVariant{
				ID:        5,
				ProductID: 1,
				Size:      "24x36",
				Price:     129.99,
			},
			Quantity: 2,
		},
	})
	if err != nil {
		t.Fatalf("createStripeCheckoutSession returned error: %v", err)
	}

	if form.Get("automatic_tax[enabled]") != "true" {
		t.Fatalf("expected automatic tax enabled, got %q", form.Get("automatic_tax[enabled]"))
	}
	if form.Get("line_items[0][price_data][tax_behavior]") != "exclusive" {
		t.Fatalf("expected exclusive tax behavior, got %q", form.Get("line_items[0][price_data][tax_behavior]"))
	}
	if form.Get("line_items[0][price_data][product_data][tax_code]") != "txcd_test_art" {
		t.Fatalf("expected configured tax code, got %q", form.Get("line_items[0][price_data][product_data][tax_code]"))
	}
	if form.Get("line_items[0][price_data][unit_amount]") != "12999" {
		t.Fatalf("expected unit amount in cents, got %q", form.Get("line_items[0][price_data][unit_amount]"))
	}
}

func TestStripeOrderAmountsFromSession(t *testing.T) {
	var session stripeCheckoutSession
	body := []byte(`{
		"amount_subtotal": 12999,
		"amount_total": 14689,
		"total_details": {
			"amount_shipping": 500,
			"amount_tax": 1190
		}
	}`)

	if err := json.Unmarshal(body, &session); err != nil {
		t.Fatalf("unmarshal stripe checkout session: %v", err)
	}

	amounts := stripeOrderAmountsFromSession(session)
	if amounts.Subtotal != 129.99 {
		t.Fatalf("expected subtotal 129.99, got %v", amounts.Subtotal)
	}
	if amounts.TaxAmount != 11.9 {
		t.Fatalf("expected tax 11.9, got %v", amounts.TaxAmount)
	}
	if amounts.ShippingAmount != 5 {
		t.Fatalf("expected shipping 5, got %v", amounts.ShippingAmount)
	}
	if amounts.TotalAmount != 146.89 {
		t.Fatalf("expected total 146.89, got %v", amounts.TotalAmount)
	}
}

func TestTrackOrderRequiresOrderNumberAndEmail(t *testing.T) {
	customerEmail := "buyer@example.com"
	store := newFakeOrderStore()
	store.ordersByNumber["ORD-TRACK"] = model.Order{
		ID:            1,
		UserID:        10,
		OrderNumber:   "ORD-TRACK",
		CustomerEmail: &customerEmail,
	}
	orderService := NewOrderService(config.Config{}, nil, store)

	order, err := orderService.TrackOrder(context.Background(), model.TrackOrderRequest{
		OrderNumber: "ORD-TRACK",
		Email:       " Buyer@Example.com ",
	})
	if err != nil {
		t.Fatalf("TrackOrder returned error: %v", err)
	}
	if order.OrderNumber != "ORD-TRACK" {
		t.Fatalf("expected order ORD-TRACK, got %q", order.OrderNumber)
	}

	_, err = orderService.TrackOrder(context.Background(), model.TrackOrderRequest{
		OrderNumber: "ORD-TRACK",
		Email:       "wrong@example.com",
	})
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound for wrong email, got %v", err)
	}
}

func TestGetByStripeSessionIDRequiresCurrentUser(t *testing.T) {
	store := newFakeOrderStore()
	store.ordersBySession["cs_test"] = model.Order{
		ID:              1,
		UserID:          10,
		OrderNumber:     "ORD-SESSION",
		StripeSessionID: stringPointer("cs_test"),
	}
	orderService := NewOrderService(config.Config{}, nil, store)

	order, err := orderService.GetByStripeSessionID(context.Background(), 10, "cs_test")
	if err != nil {
		t.Fatalf("GetByStripeSessionID returned error: %v", err)
	}
	if order.OrderNumber != "ORD-SESSION" {
		t.Fatalf("expected ORD-SESSION, got %q", order.OrderNumber)
	}

	_, err = orderService.GetByStripeSessionID(context.Background(), 20, "cs_test")
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound for another user, got %v", err)
	}
}

func TestHandleStripeWebhookSendsOrderConfirmation(t *testing.T) {
	customerEmail := "buyer@example.com"
	store := newFakeOrderStore()
	store.ordersBySession["cs_paid"] = model.Order{
		ID:              1,
		UserID:          10,
		OrderNumber:     "ORD-PAID",
		Currency:        "USD",
		StripeSessionID: stringPointer("cs_paid"),
		CustomerEmail:   &customerEmail,
		Items: []model.OrderItem{
			{ProductTitle: "Canvas", Quantity: 1, Subtotal: 129.99},
		},
	}
	carts := &fakeCartStore{}
	emails := &fakeOrderEmailSender{}
	orderService := NewOrderService(config.Config{StripeWebhookSecret: "whsec_test"}, carts, store)
	orderService.emails = emails

	payload := signedStripePayload(t, "whsec_test", map[string]any{
		"id":   "evt_paid",
		"type": "checkout.session.completed",
		"data": map[string]any{
			"object": map[string]any{
				"id":              "cs_paid",
				"payment_intent":  "pi_test",
				"payment_status":  "paid",
				"amount_subtotal": 12999,
				"amount_total":    14129,
				"total_details": map[string]any{
					"amount_tax":      1130,
					"amount_shipping": 0,
				},
				"customer_details": map[string]any{
					"email": "buyer@example.com",
					"name":  "Buyer",
				},
			},
		},
	})

	if err := orderService.HandleStripeWebhook(context.Background(), payload.body, payload.signature); err != nil {
		t.Fatalf("HandleStripeWebhook returned error: %v", err)
	}
	if len(emails.sent) != 1 || emails.sent[0].OrderNumber != "ORD-PAID" {
		t.Fatalf("expected one confirmation email for ORD-PAID, got %#v", emails.sent)
	}
	if carts.clearedCartID != 10 {
		t.Fatalf("expected cart clear for user/cart 10, got %d", carts.clearedCartID)
	}
	updated := store.ordersBySession["cs_paid"]
	if updated.TaxAmount != 11.3 || updated.TotalAmount != 141.29 {
		t.Fatalf("expected Stripe totals to be saved, got tax=%v total=%v", updated.TaxAmount, updated.TotalAmount)
	}
}

func TestHandleStripeWebhookIgnoresEmailFailure(t *testing.T) {
	customerEmail := "buyer@example.com"
	store := newFakeOrderStore()
	store.ordersBySession["cs_email_error"] = model.Order{
		ID:              1,
		UserID:          10,
		OrderNumber:     "ORD-EMAIL-ERROR",
		StripeSessionID: stringPointer("cs_email_error"),
		CustomerEmail:   &customerEmail,
	}
	carts := &fakeCartStore{}
	orderService := NewOrderService(config.Config{StripeWebhookSecret: "whsec_test"}, carts, store)
	orderService.emails = &fakeOrderEmailSender{err: errors.New("smtp unavailable")}

	payload := signedStripePayload(t, "whsec_test", map[string]any{
		"id":   "evt_paid",
		"type": "checkout.session.completed",
		"data": map[string]any{
			"object": map[string]any{
				"id":               "cs_email_error",
				"payment_status":   "paid",
				"amount_subtotal":  1000,
				"amount_total":     1000,
				"customer_details": map[string]any{"email": "buyer@example.com"},
			},
		},
	})

	if err := orderService.HandleStripeWebhook(context.Background(), payload.body, payload.signature); err != nil {
		t.Fatalf("expected webhook to ignore email failure, got %v", err)
	}
	if carts.clearedCartID != 10 {
		t.Fatalf("expected cart clear after email failure, got %d", carts.clearedCartID)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type signedPayload struct {
	body      []byte
	signature string
}

func signedStripePayload(t *testing.T, secret string, payload map[string]any) signedPayload {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := "t=" + timestamp + ",v1=" + computeStripeSignature(timestamp, body, secret)

	return signedPayload{body: body, signature: signature}
}

type fakeOrderStore struct {
	ordersBySession map[string]model.Order
	ordersByNumber  map[string]model.Order
}

func newFakeOrderStore() *fakeOrderStore {
	return &fakeOrderStore{
		ordersBySession: make(map[string]model.Order),
		ordersByNumber:  make(map[string]model.Order),
	}
}

func (s *fakeOrderStore) Create(_ context.Context, request model.CreateOrderRequest) (model.Order, error) {
	return model.Order{ID: 1, UserID: request.UserID, OrderNumber: request.OrderNumber}, nil
}

func (s *fakeOrderStore) FindAllByUser(_ context.Context, _ int64) ([]model.OrderSummary, error) {
	return nil, nil
}

func (s *fakeOrderStore) FindByID(_ context.Context, userID int64, id int64) (model.Order, error) {
	for _, order := range s.ordersBySession {
		if order.UserID == userID && order.ID == id {
			return order, nil
		}
	}
	return model.Order{}, sql.ErrNoRows
}

func (s *fakeOrderStore) FindByStripeSessionID(_ context.Context, stripeSessionID string) (model.Order, error) {
	order, ok := s.ordersBySession[stripeSessionID]
	if !ok {
		return model.Order{}, sql.ErrNoRows
	}
	return order, nil
}

func (s *fakeOrderStore) FindByStripeSessionIDForUser(_ context.Context, userID int64, stripeSessionID string) (model.Order, error) {
	order, ok := s.ordersBySession[stripeSessionID]
	if !ok || order.UserID != userID {
		return model.Order{}, sql.ErrNoRows
	}
	return order, nil
}

func (s *fakeOrderStore) FindByOrderNumberAndCustomerEmail(_ context.Context, orderNumber string, email string) (model.Order, error) {
	order, ok := s.ordersByNumber[orderNumber]
	if !ok || order.CustomerEmail == nil || strings.ToLower(*order.CustomerEmail) != strings.ToLower(email) {
		return model.Order{}, sql.ErrNoRows
	}
	return order, nil
}

func (s *fakeOrderStore) UpdateStripeSession(_ context.Context, orderID int64, stripeSessionID string) error {
	order := model.Order{ID: orderID, StripeSessionID: &stripeSessionID}
	s.ordersBySession[stripeSessionID] = order
	return nil
}

func (s *fakeOrderStore) MarkPaidByStripeSessionID(_ context.Context, stripeSessionID string, paymentIntentID string, paymentStatus string, subtotal float64, taxAmount float64, shippingAmount float64, totalAmount float64, customerEmail *string, customerName *string, shippingName *string, shippingPhone *string, shippingLine1 *string, shippingLine2 *string, shippingCity *string, shippingState *string, shippingPostalCode *string, shippingCountry *string) (bool, error) {
	order, ok := s.ordersBySession[stripeSessionID]
	if !ok {
		return false, nil
	}
	order.Status = "paid"
	order.PaymentStatus = paymentStatus
	order.StripePaymentIntentID = stringPointer(paymentIntentID)
	order.Subtotal = subtotal
	order.TaxAmount = taxAmount
	order.ShippingAmount = shippingAmount
	order.TotalAmount = totalAmount
	order.CustomerEmail = customerEmail
	order.CustomerName = customerName
	order.ShippingName = shippingName
	order.ShippingPhone = shippingPhone
	order.ShippingLine1 = shippingLine1
	order.ShippingLine2 = shippingLine2
	order.ShippingCity = shippingCity
	order.ShippingState = shippingState
	order.ShippingPostalCode = shippingPostalCode
	order.ShippingCountry = shippingCountry
	s.ordersBySession[stripeSessionID] = order
	s.ordersByNumber[order.OrderNumber] = order
	return true, nil
}

func (s *fakeOrderStore) MarkExpiredByStripeSessionID(_ context.Context, _ string) error {
	return nil
}

func (s *fakeOrderStore) MarkCheckoutFailed(_ context.Context, _ int64) error {
	return nil
}

type fakeCartStore struct {
	clearedCartID int64
}

func (s *fakeCartStore) GetOrCreate(_ context.Context, userID int64) (model.Cart, error) {
	return model.Cart{ID: userID, UserID: userID}, nil
}

func (s *fakeCartStore) FindItems(_ context.Context, _ int64) ([]model.CartItem, error) {
	return nil, nil
}

func (s *fakeCartStore) Clear(_ context.Context, cartID int64) error {
	s.clearedCartID = cartID
	return nil
}

type fakeOrderEmailSender struct {
	sent []model.Order
	err  error
}

func (s *fakeOrderEmailSender) SendOrderConfirmation(_ context.Context, order model.Order) error {
	s.sent = append(s.sent, order)
	return s.err
}

func stringPointer(value string) *string {
	return &value
}
