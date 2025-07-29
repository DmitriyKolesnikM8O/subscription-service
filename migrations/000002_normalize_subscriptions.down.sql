ALTER TABLE subscriptions
ADD COLUMN service_name TEXT,
ADD COLUMN price INTEGER;

UPDATE subscriptions
SET service_name = services.name,
    price = services.price
FROM services
WHERE subscriptions.service_id = services.id;

ALTER TABLE subscriptions
ALTER COLUMN service_name SET NOT NULL,
ALTER COLUMN price SET NOT NULL;

ALTER TABLE subscriptions
DROP CONSTRAINT fk_service;

DROP INDEX idx_subscriptions_service_id;

ALTER TABLE subscriptions
DROP COLUMN service_id;

DROP TABLE services; 