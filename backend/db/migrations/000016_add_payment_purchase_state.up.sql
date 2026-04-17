CREATE TABLE app.user_payment_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    provider TEXT NOT NULL CHECK (
        provider IN ('ccbill')
    ),
    provider_payment_token_ref TEXT NOT NULL UNIQUE,
    provider_payment_account_ref TEXT NOT NULL,
    brand TEXT NOT NULL CHECK (
        brand IN ('visa', 'mastercard', 'jcb', 'american_express')
    ),
    last4 TEXT NOT NULL CHECK (
        last4 ~ '^[0-9]{4}$'
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id, user_id),
    UNIQUE (user_id, provider_payment_account_ref)
);

CREATE INDEX idx_user_payment_methods_user_last_used_at
    ON app.user_payment_methods (user_id, last_used_at DESC, id DESC);

CREATE TABLE app.main_purchase_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    main_id UUID NOT NULL REFERENCES app.mains (id) ON DELETE CASCADE,
    from_short_id UUID NOT NULL REFERENCES app.shorts (id),
    provider TEXT NOT NULL CHECK (
        provider IN ('ccbill')
    ),
    payment_method_mode TEXT NOT NULL CHECK (
        payment_method_mode IN ('saved_card', 'new_card')
    ),
    user_payment_method_id UUID,
    provider_payment_token_ref TEXT NOT NULL,
    idempotency_key TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK (
        status IN ('processing', 'succeeded', 'pending', 'failed')
    ),
    failure_reason TEXT CHECK (
        failure_reason IS NULL
        OR failure_reason IN (
            'card_brand_unsupported',
            'purchase_declined',
            'authentication_failed'
        )
    ),
    pending_reason TEXT CHECK (
        pending_reason IS NULL
        OR pending_reason IN ('provider_processing')
    ),
    provider_purchase_ref TEXT UNIQUE,
    provider_transaction_ref TEXT,
    provider_session_ref TEXT,
    provider_payment_unique_ref TEXT,
    provider_decline_code INTEGER,
    provider_decline_text TEXT,
    requested_price_jpy BIGINT NOT NULL CHECK (
        requested_price_jpy > 0
    ),
    requested_currency_code INTEGER NOT NULL CHECK (
        requested_currency_code BETWEEN 1 AND 999
    ),
    accepted_age BOOLEAN NOT NULL DEFAULT FALSE,
    accepted_terms BOOLEAN NOT NULL DEFAULT FALSE,
    provider_processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_payment_method_id, user_id)
        REFERENCES app.user_payment_methods (id, user_id),
    CHECK (
        (status = 'failed' AND failure_reason IS NOT NULL)
        OR (status <> 'failed' AND failure_reason IS NULL)
    ),
    CHECK (
        (status = 'pending' AND pending_reason IS NOT NULL)
        OR (status <> 'pending' AND pending_reason IS NULL)
    ),
    CHECK (
        (status = 'processing' AND provider_processed_at IS NULL)
        OR (status <> 'processing' AND provider_processed_at IS NOT NULL)
    )
);

CREATE INDEX idx_main_purchase_attempts_user_main_created_at
    ON app.main_purchase_attempts (user_id, main_id, created_at DESC, id DESC);

CREATE INDEX idx_main_purchase_attempts_provider_purchase_ref
    ON app.main_purchase_attempts (provider_purchase_ref)
    WHERE provider_purchase_ref IS NOT NULL;

CREATE UNIQUE INDEX idx_main_purchase_attempts_user_main_inflight
    ON app.main_purchase_attempts (user_id, main_id)
    WHERE status IN ('processing', 'pending');
