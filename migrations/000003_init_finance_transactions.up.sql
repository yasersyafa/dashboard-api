CREATE TYPE transaction_type AS ENUM ('INCOME', 'EXPENSE');

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type transaction_type NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    note VARCHAR(140),
    occurred_at TIMESTAMPTZ NOT NULL,
    category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_transactions_category_id ON transactions (category_id);
CREATE INDEX idx_transactions_occurred_at ON transactions (occurred_at);