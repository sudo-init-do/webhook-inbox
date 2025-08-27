package models

import (
	"time"

	"github.com/google/uuid"
)

type Provider string

const (
	ProviderStripe     Provider = "stripe"
	ProviderFlutterwave         = "flutterwave"
	ProviderPaystack            = "paystack"
	ProviderGitHub              = "github"
)

type Endpoint struct {
	ID        int64
	Token     uuid.UUID
	Provider  Provider
	Secret    string
	CreatedAt time.Time
}

type Message struct {
	ID         int64
	EndpointID int64
	HeadersRaw []byte // stored as JSON
	Body       string
	ReceivedAt time.Time
}

type Replay struct {
	ID         int64
	MessageID  int64
	TargetURL  string
	StatusCode *int
	RespBody   *string
	CreatedAt  time.Time
}
