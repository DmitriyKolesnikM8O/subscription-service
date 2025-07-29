CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0)
);

INSERT INTO services (name, price)
SELECT DISTINCT service_name, price
FROM subscriptions;

ALTER TABLE subscriptions
ADD COLUMN service_id UUID;

UPDATE subscriptions
SET service_id = (SELECT id FROM services WHERE services.name = subscriptions.service_name AND services.price = subscriptions.price);

ALTER TABLE subscriptions
ALTER COLUMN service_id SET NOT NULL;

ALTER TABLE subscriptions
ADD CONSTRAINT fk_service
FOREIGN KEY (service_id)
REFERENCES services(id)
ON DELETE RESTRICT;

CREATE INDEX idx_subscriptions_service_id ON subscriptions(service_id);

ALTER TABLE subscriptions
DROP COLUMN service_name,
DROP COLUMN price; 