CREATE TABLE IF NOT EXISTS endpoints (
  id BIGSERIAL PRIMARY KEY,
  token UUID UNIQUE NOT NULL,
  provider TEXT NOT NULL CHECK (provider IN ('stripe','flutterwave','paystack','github')),
  secret TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS messages (
  id BIGSERIAL PRIMARY KEY,
  endpoint_id BIGINT NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
  headers_json JSONB NOT NULL,
  body TEXT NOT NULL,
  received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_endpoint_id ON messages(endpoint_id);

CREATE TABLE IF NOT EXISTS replays (
  id BIGSERIAL PRIMARY KEY,
  message_id BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
  target_url TEXT NOT NULL,
  status_code INT,
  response_body TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
