-- 000002_normalize_subscriptions.down.sql

-- 1. Восстанавливаем удаленные колонки (если они были удалены)
ALTER TABLE subscriptions
ADD COLUMN IF NOT EXISTS service_name TEXT,
ADD COLUMN IF NOT EXISTS price INTEGER;

-- 2. Заполняем их значениями из таблицы services (если она существует)
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

-- 3. Удаляем внешний ключ (если он существует)
ALTER TABLE subscriptions
DROP CONSTRAINT IF EXISTS fk_service;

-- 4. Удаляем индекс по service_id (если он существует)
DROP INDEX IF EXISTS idx_subscriptions_service_id;

-- 5. Удаляем колонку service_id (если она существует)
ALTER TABLE subscriptions
DROP COLUMN IF EXISTS service_id;

-- 6. Удаляем таблицу services (если она существует)
DROP TABLE IF EXISTS services;