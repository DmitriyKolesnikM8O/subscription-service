CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    service_name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_service ON subscriptions(service_name);
