package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/google/uuid"
	"github.com/you/webhook-inbox/internal/models"
)

type Store struct {
	DB *sql.DB
}

func Open(dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	return &Store{DB: db}, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) Migrate(ctx context.Context, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sqlBytes, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, string(sqlBytes))
	return err
}

func (s *Store) CreateEndpoint(ctx context.Context, provider models.Provider, secret string) (models.Endpoint, error) {
	tok := uuid.New()
	var ep models.Endpoint
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO endpoints (token, provider, secret)
		VALUES ($1,$2,$3)
		RETURNING id, token, provider, secret, created_at
	`, tok, provider, secret).Scan(&ep.ID, &ep.Token, &ep.Provider, &ep.Secret, &ep.CreatedAt)
	return ep, err
}

func (s *Store) GetEndpointByToken(ctx context.Context, token uuid.UUID) (models.Endpoint, error) {
	var ep models.Endpoint
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, token, provider, secret, created_at
		FROM endpoints
		WHERE token=$1
	`, token).Scan(&ep.ID, &ep.Token, &ep.Provider, &ep.Secret, &ep.CreatedAt)
	return ep, err
}

func (s *Store) InsertMessage(ctx context.Context, endpointID int64, headers map[string][]string, body []byte) (models.Message, error) {
	h, _ := json.Marshal(headers)
	var m models.Message
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO messages (endpoint_id, headers_json, body)
		VALUES ($1, $2, $3)
		RETURNING id, endpoint_id, headers_json, body, received_at
	`, endpointID, h, string(body)).Scan(&m.ID, &m.EndpointID, &m.HeadersRaw, &m.Body, &m.ReceivedAt)
	return m, err
}

func (s *Store) ListMessages(ctx context.Context, endpointID *int64, limit int) ([]models.Message, error) {
	query := `
		SELECT id, endpoint_id, headers_json, body, received_at
		FROM messages
	`
	var rows *sql.Rows
	var err error
	if endpointID != nil {
		query += ` WHERE endpoint_id=$1 ORDER BY id DESC LIMIT $2`
		rows, err = s.DB.QueryContext(ctx, query, *endpointID, limit)
	} else {
		query += ` ORDER BY id DESC LIMIT $1`
		rows, err = s.DB.QueryContext(ctx, query, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.EndpointID, &m.HeadersRaw, &m.Body, &m.ReceivedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) GetMessage(ctx context.Context, id int64) (models.Message, error) {
	var m models.Message
	err := s.DB.QueryRowContext(ctx, `
		SELECT id, endpoint_id, headers_json, body, received_at
		FROM messages WHERE id=$1
	`, id).Scan(&m.ID, &m.EndpointID, &m.HeadersRaw, &m.Body, &m.ReceivedAt)
	return m, err
}

func (s *Store) InsertReplay(ctx context.Context, msgID int64, targetURL string, statusCode int, respBody []byte) (models.Replay, error) {
	var r models.Replay
	err := s.DB.QueryRowContext(ctx, `
		INSERT INTO replays (message_id, target_url, status_code, response_body)
		VALUES ($1,$2,$3,$4)
		RETURNING id, message_id, target_url, status_code, response_body, created_at
	`, msgID, targetURL, statusCode, string(respBody)).Scan(&r.ID, &r.MessageID, &r.TargetURL, &r.StatusCode, &r.RespBody, &r.CreatedAt)
	return r, err
}

func (s *Store) MustURL(base string, token uuid.UUID) string {
	return fmt.Sprintf("%s/hooks/%s", base, token.String())
}
