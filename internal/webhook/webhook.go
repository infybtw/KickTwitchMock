package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	EventTypes []string  `json:"event_types"`
	Secret     string    `json:"secret,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Sender struct {
	mu            sync.RWMutex
	subscriptions map[string]Subscription
	logger        *slog.Logger
	client        *http.Client
}

func New(logger *slog.Logger) *Sender {
	return &Sender{
		subscriptions: make(map[string]Subscription),
		logger:        logger,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Sender) CreateSubscription(url string, eventTypes []string, secret string) Subscription {
	if secret == "" {
		secret = uuid.NewString()
	}

	sub := Subscription{
		ID:         uuid.NewString(),
		URL:        url,
		EventTypes: eventTypes,
		Secret:     secret,
		CreatedAt:  time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.subscriptions[sub.ID] = sub

	return sub
}

func (s *Sender) DeleteSubscription(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.subscriptions[id]; !exists {
		return false
	}

	delete(s.subscriptions, id)
	return true
}

func (s *Sender) ListSubscriptions() []Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Subscription, 0, len(s.subscriptions))
	for _, sub := range s.subscriptions {
		result = append(result, sub)
	}

	return result
}

func (s *Sender) Broadcast(eventType string, eventData map[string]any) {
	s.mu.RLock()
	subs := make([]Subscription, 0)
	for _, sub := range s.subscriptions {
		for _, t := range sub.EventTypes {
			if t == eventType || t == "*" {
				subs = append(subs, sub)
				break
			}
		}
	}
	s.mu.RUnlock()

	for _, sub := range subs {
		go s.deliver(sub, eventType, eventData)
	}
}

func (s *Sender) deliver(sub Subscription, eventType string, eventData map[string]any) {
	body, err := json.Marshal(eventData)
	if err != nil {
		s.logger.Error("failed to marshal webhook payload",
			slog.String("subscription_id", sub.ID),
			slog.Any("error", err),
		)
		return
	}

	req, err := http.NewRequest(http.MethodPost, sub.URL, bytes.NewReader(body))
	if err != nil {
		s.logger.Error("failed to create webhook request",
			slog.String("subscription_id", sub.ID),
			slog.String("url", sub.URL),
			slog.Any("error", err),
		)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Kick-Event-Type", eventType)
	req.Header.Set("Kick-Event-Version", "1")
	req.Header.Set("X-Kick-Signature", signPayload(body, sub.Secret))

	resp, err := s.client.Do(req)
	if err != nil {
		s.logger.Error("webhook delivery failed",
			slog.String("subscription_id", sub.ID),
			slog.String("url", sub.URL),
			slog.String("event", eventType),
			slog.Any("error", err),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		s.logger.Warn("webhook returned error status",
			slog.String("subscription_id", sub.ID),
			slog.String("url", sub.URL),
			slog.String("event", eventType),
			slog.Int("status", resp.StatusCode),
		)
		return
	}

	s.logger.Info("webhook delivered",
		slog.String("subscription_id", sub.ID),
		slog.String("url", sub.URL),
		slog.String("event", eventType),
		slog.Int("status", resp.StatusCode),
	)
}

func signPayload(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
