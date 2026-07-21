CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    tier_id UUID NOT NULL REFERENCES tiers(id),
    tier_name VARCHAR(32) NOT NULL,
    tier_level SMALLINT NOT NULL,
    tier_price NUMERIC(14, 2) NOT NULL,
    tier_currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    status SMALLINT NOT NULL CHECK (status IN (1, 2, 3)),
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX subscriptions_one_active_user_unique
    ON subscriptions(user_id)
    WHERE status = 1;
