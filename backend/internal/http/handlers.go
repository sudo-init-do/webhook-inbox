package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/you/webhook-inbox/internal/config"
	"github.com/you/webhook-inbox/internal/models"
	"github.com/you/webhook-inbox/internal/providers"
	"github.com/you/webhook-inbox/internal/storage"
)

type Handlers struct {
	Store  *storage.Store
	Config config.Config
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

type createEndpointReq struct {
	Provider string `json:"provider"` // stripe|flutterwave|paystack|github
	Secret   string `json:"secret"`   // required for all providers
}
type createEndpointResp struct {
	ID    int64  `json:"id"`
	Token string `json:"token"`
	URL   string `json:"url"`
}

func (h *Handlers) CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var req createEndpointReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	p := models.Provider(req.Provider)
	switch p {
	case models.ProviderStripe, models.ProviderFlutterwave, models.ProviderPaystack, models.ProviderGitHub:
	default:
		http.Error(w, "unsupported provider", http.StatusBadRequest)
		return
	}
	if req.Secret == "" {
		http.Error(w, "secret required", http.StatusBadRequest)
		return
	}
	ep, err := h.Store.CreateEndpoint(r.Context(), p, req.Secret)
	if err != nil {
		http.Error(w, "failed to create endpoint", http.StatusInternalServerError)
		return
	}
	resp := createEndpointResp{
		ID:    ep.ID,
		Token: ep.Token.String(),
		URL:   h.Store.MustURL(h.Config.PublicBase, ep.Token),
	}
	writeJSON(w, resp, http.StatusCreated)
}

func (h *Handlers) ListMessages(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if q := r.URL.Query().Get("limit"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 && v <= 200 {
			limit = v
		}
	}
	var endpointID *int64
	if q := r.URL.Query().Get("endpointId"); q != "" {
		if v, err := strconv.ParseInt(q, 10, 64); err == nil {
			endpointID = &v
		}
	}
	items, err := h.Store.ListMessages(r.Context(), endpointID, limit)
	if err != nil {
		http.Error(w, "failed to list messages", http.StatusInternalServerError)
		return
	}
	writeJSON(w, items, http.StatusOK)
}

func (h *Handlers) GetMessage(w http.ResponseWriter, r *http.Request) {
	idStr := pathParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	m, err := h.Store.GetMessage(r.Context(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, m, http.StatusOK)
}

func (h *Handlers) ReceiveHook(w http.ResponseWriter, r *http.Request) {
	tokStr := pathParam(r, "token")
	tok, err := uuid.Parse(tokStr)
	if err != nil {
		http.Error(w, "bad token", http.StatusBadRequest)
		return
	}
	ep, err := h.Store.GetEndpointByToken(r.Context(), tok)
	if err != nil {
		http.Error(w, "endpoint not found", http.StatusNotFound)
		return
	}
	body, _ := io.ReadAll(r.Body)

	// Verify per provider
	switch ep.Provider {
	case models.ProviderStripe:
		if err := providers.VerifyStripeSignature(ep.Secret, r.Header.Get("Stripe-Signature"), body, 5*time.Minute); err != nil {
			http.Error(w, "invalid stripe signature: "+err.Error(), http.StatusUnauthorized)
			return
		}
	case models.ProviderFlutterwave:
		if !providers.VerifyFlutterwaveSignature(ep.Secret, r.Header.Get("verif-hash")) {
			http.Error(w, "invalid flutterwave signature", http.StatusUnauthorized)
			return
		}
	case models.ProviderPaystack:
		if !providers.VerifyPaystackSignature(ep.Secret, r.Header.Get("x-paystack-signature")) {
			http.Error(w, "invalid paystack signature", http.StatusUnauthorized)
			return
		}
	case models.ProviderGitHub:
		if err := providers.VerifyGitHubSignature(ep.Secret, r.Header.Get("X-Hub-Signature-256"), body); err != nil {
			http.Error(w, "invalid github signature: "+err.Error(), http.StatusUnauthorized)
			return
		}
	default:
		http.Error(w, "unsupported provider", http.StatusBadRequest)
		return
	}

	_, err = h.Store.InsertMessage(r.Context(), ep.ID, r.Header, body)
	if err != nil {
		http.Error(w, "failed to store message", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("received"))
}

// ===== Replay support =====

type replayReq struct {
	TargetURL string `json:"target_url"`
}

func (h *Handlers) ReplayMessage(w http.ResponseWriter, r *http.Request) {
	idStr := pathParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	var req replayReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TargetURL == "" {
		http.Error(w, "target_url required", http.StatusBadRequest)
		return
	}

	// 1) Load original message
	msg, err := h.Store.GetMessage(r.Context(), id)
	if err != nil {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}

	// 2) Rebuild a minimal header set (preserve content-type if present)
	var hdr map[string][]string
	_ = json.Unmarshal(msg.HeadersRaw, &hdr)

	forward := http.Header{}
	if v := hdr["Content-Type"]; len(v) > 0 {
		forward.Set("Content-Type", v[0])
	} else {
		forward.Set("Content-Type", "application/json")
	}
	forward.Set("X-Replayed-From", "webhook-inbox")

	// 3) Send
	client := &http.Client{Timeout: 10 * time.Second}
	reqOut, err := http.NewRequestWithContext(r.Context(), http.MethodPost, req.TargetURL, bytes.NewBufferString(msg.Body))
	if err != nil {
		http.Error(w, "build replay request failed", http.StatusInternalServerError)
		return
	}
	reqOut.Header = forward

	resp, err := client.Do(reqOut)
	if err != nil {
		http.Error(w, "replay failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	// 4) Store replay result
	_, _ = h.Store.InsertReplay(r.Context(), msg.ID, req.TargetURL, resp.StatusCode, respBody)

	// 5) Return summary
	writeJSON(w, map[string]any{
		"message_id": msg.ID,
		"target_url": req.TargetURL,
		"status":     resp.StatusCode,
		"response":   string(respBody),
	}, http.StatusOK)
}

/*** helpers ***/
func writeJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
