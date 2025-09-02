-- Создание таблиц для логирования платежей

-- Таблица payments - основная информация о платежах
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    booking_id BIGINT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    payment_id VARCHAR(255) NOT NULL UNIQUE, -- ID от FreedomPay
    amount INTEGER NOT NULL, -- сумма в тенге
    currency VARCHAR(10) NOT NULL DEFAULT 'KZT',
    status VARCHAR(50) NOT NULL, -- pending, processing, success, failed, expired
    payment_method VARCHAR(100), -- способ оплаты от FreedomPay
    provider_status VARCHAR(50), -- оригинальный статус от FreedomPay
    provider_response JSONB, -- полный ответ от FreedomPay
    final_booking_status VARCHAR(20), -- итоговый статус бронирования после оплаты
    processed_at TIMESTAMPTZ, -- время успешной обработки
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_payments_status CHECK (status IN ('pending', 'processing', 'success', 'failed', 'expired', 'canceled')),
    CONSTRAINT chk_payments_amount CHECK (amount > 0)
);

-- Таблица payment_logs - детальное логирование всех проверок статуса
CREATE TABLE payment_logs (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT REFERENCES payments(id) ON DELETE CASCADE, -- может быть NULL для неуспешных созданий
    booking_id BIGINT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    fp_payment_id VARCHAR(255) NOT NULL, -- ID от FreedomPay
    action VARCHAR(50) NOT NULL, -- create_payment, check_status, process_payment
    old_status VARCHAR(50), -- предыдущий статус
    new_status VARCHAR(50), -- новый статус  
    fp_response JSONB, -- ответ от FreedomPay API
    processing_duration INTEGER, -- время выполнения запроса в миллисекундах
    user_id BIGINT REFERENCES users(id), -- кто инициировал действие
    source VARCHAR(20) NOT NULL DEFAULT 'api', -- api, webhook, manual, scheduler
    success BOOLEAN NOT NULL DEFAULT false, -- успешно ли выполнено действие
    error_message TEXT, -- сообщение об ошибке если есть
    ip_address INET, -- IP адрес пользователя
    user_agent TEXT, -- User-Agent браузера
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_payment_logs_action CHECK (action IN ('create_payment', 'check_status', 'process_payment', 'webhook_notification'))
);

-- Индексы для быстрого поиска
CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_payment_id ON payments(payment_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_created_at ON payments(created_at);
CREATE INDEX idx_payments_processed_at ON payments(processed_at) WHERE processed_at IS NOT NULL;

CREATE INDEX idx_payment_logs_payment_id ON payment_logs(payment_id) WHERE payment_id IS NOT NULL;
CREATE INDEX idx_payment_logs_booking_id ON payment_logs(booking_id);
CREATE INDEX idx_payment_logs_fp_payment_id ON payment_logs(fp_payment_id);
CREATE INDEX idx_payment_logs_action ON payment_logs(action);
CREATE INDEX idx_payment_logs_created_at ON payment_logs(created_at);
CREATE INDEX idx_payment_logs_user_id ON payment_logs(user_id) WHERE user_id IS NOT NULL;

-- Триггер для обновления updated_at в payments
CREATE OR REPLACE FUNCTION update_payments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_payments_updated_at 
    BEFORE UPDATE ON payments
    FOR EACH ROW 
    EXECUTE FUNCTION update_payments_updated_at();

-- Комментарии для документации
COMMENT ON TABLE payments IS 'Основная информация о платежах через FreedomPay';
COMMENT ON TABLE payment_logs IS 'Детальное логирование всех операций с платежами для аудита';

COMMENT ON COLUMN payments.payment_id IS 'Уникальный ID платежа от FreedomPay';
COMMENT ON COLUMN payments.provider_response IS 'Полный JSON ответ от FreedomPay для отладки';
COMMENT ON COLUMN payments.final_booking_status IS 'Статус бронирования после обработки платежа';

COMMENT ON COLUMN payment_logs.action IS 'Тип действия: создание, проверка статуса, обработка';
COMMENT ON COLUMN payment_logs.processing_duration IS 'Время выполнения запроса к FreedomPay в миллисекундах';
COMMENT ON COLUMN payment_logs.source IS 'Источник операции: API, webhook, ручная операция'; 