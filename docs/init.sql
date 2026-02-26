-- ==========================================================
-- ПЛАТФОРМА КЛИМАТ-КОНТРОЛЯ ТЕРРАРИУМА - СХЕМА POSTGRESQL
-- ==========================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица для настроек автоматизации и климата
CREATE TABLE IF NOT EXISTS automation_settings (
    id SERIAL PRIMARY KEY,
    warm_target_min NUMERIC(5, 2) NOT NULL DEFAULT 31.0,
    warm_target_max NUMERIC(5, 2) NOT NULL DEFAULT 33.0,
    cold_max_threshold NUMERIC(5, 2) NOT NULL DEFAULT 26.0,
    emergency_max_threshold NUMERIC(5, 2) NOT NULL DEFAULT 35.0,
    humidity_min NUMERIC(5, 2) NOT NULL DEFAULT 50.0,
    humidity_max NUMERIC(5, 2) NOT NULL DEFAULT 65.0,
    hysteresis_temp NUMERIC(4, 2) NOT NULL DEFAULT 0.5,
    hysteresis_hum NUMERIC(4, 2) NOT NULL DEFAULT 2.0,
    mode VARCHAR(20) NOT NULL DEFAULT 'AUTO', -- 'AUTO' или 'MANUAL'
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Убедимся, что существует только одна строка активной конфигурации
CREATE UNIQUE INDEX IF NOT EXISTS single_config_idx ON automation_settings((1));
INSERT INTO automation_settings (id) VALUES (1) ON CONFLICT DO NOTHING;

-- Таблица для расписания освещения и других задач
CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    relay_id VARCHAR(50) NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Временные ряды большого объема: Показания датчиков
CREATE TABLE IF NOT EXISTS sensor_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    warm_zone_temp NUMERIC(5, 2),
    warm_zone_hum NUMERIC(5, 2),
    cold_zone_temp NUMERIC(5, 2),
    cold_zone_hum NUMERIC(5, 2)
);

-- Индекс для запросов временных рядов
CREATE INDEX IF NOT EXISTS idx_sensor_logs_time ON sensor_logs(recorded_at DESC);

-- Отслеживание включения/выключения реле для восстановления состояния и расчета потребления
CREATE TABLE IF NOT EXISTS relay_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    relay_id VARCHAR(50) NOT NULL,
    state BOOLEAN NOT NULL,
    reason VARCHAR(100), -- 'AUTO_TEMP_TRIGGER', 'MANUAL_OVERRIDE', 'EMERGENCY_CUTOFF'
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_relay_logs_time ON relay_logs(recorded_at DESC);

-- Агрегированные отчеты об энергопотреблении (например, генерируемые ежедневно)
CREATE TABLE IF NOT EXISTS energy_reports (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    report_date DATE NOT NULL UNIQUE,
    heat_mat_kwh NUMERIC(8, 4) DEFAULT 0,
    light_kwh NUMERIC(8, 4) DEFAULT 0,
    fogger_kwh NUMERIC(8, 4) DEFAULT 0,
    spare_kwh NUMERIC(8, 4) DEFAULT 0,
    total_kwh NUMERIC(8, 4) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Минимальное системное состояние для восстановления (Полезно вместе с JSON для быстрого восстановления после сбоя)
CREATE TABLE IF NOT EXISTS system_state (
    id SERIAL PRIMARY KEY,
    heat_mat BOOLEAN DEFAULT false,
    fogger BOOLEAN DEFAULT false,
    light BOOLEAN DEFAULT false,
    spare BOOLEAN DEFAULT false,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS single_state_idx ON system_state((1));
INSERT INTO system_state (id) VALUES (1) ON CONFLICT DO NOTHING;
