ALTER TABLE subscriptions
ADD COLUMN IF NOT EXISTS service_name TEXT,
ADD COLUMN IF NOT EXISTS price INTEGER;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'services') THEN
        UPDATE subscriptions
        SET service_name = services.name,
            price = services.price
        FROM services
        WHERE subscriptions.service_id = services.id;
    END IF;
END $$;

ALTER TABLE subscriptions
DROP CONSTRAINT IF EXISTS fk_service;


DROP INDEX IF EXISTS idx_subscriptions_service_id;


ALTER TABLE subscriptions
DROP COLUMN IF EXISTS service_id;


DROP TABLE IF EXISTS services;