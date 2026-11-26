-- 007_fraud_reports_and_user_profile.sql
-- Dolandırıcılık bildirimleri ve kullanıcı yönetimi

CREATE TABLE IF NOT EXISTS fraud_reports (
    id BIGSERIAL PRIMARY KEY,
    reported_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reporter_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT NOT NULL DEFAULT 'Dolandırıcılık / Spam Şüphesi',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_report_reporter_target UNIQUE (reporter_user_id, reported_user_id)
);

CREATE INDEX IF NOT EXISTS idx_fraud_reports_reported_user ON fraud_reports (reported_user_id);
