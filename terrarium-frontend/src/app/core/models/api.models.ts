// Типизированные модели данных API (соответствуют Go-моделям бэкенда)

// Текущие показания датчиков (температура + влажность обеих зон)
export interface SensorCurrent {
    warm_temp: number;
    warm_hum: number;
    cold_temp: number;
    cold_hum: number;
    timestamp: string;
    mode: string;
}

// Историческая запись показаний датчиков
export interface SensorDataHistory {
    timestamp: string;
    warm_temp: number;
    warm_hum: number;
    cold_temp: number;
    cold_hum: number;
}

// Состояние 4 реле
export interface RelayState {
    heat_mat: boolean;
    fogger: boolean;
    light: boolean;
    spare: boolean;
}

// Запрос переключения реле
export interface RelayToggleRequest {
    state: boolean;
}

// Конфигурация автоматизации (пороги, гистерезис)
export interface ConfigPayload {
    warm_target_min: number;
    warm_target_max: number;
    cold_max_threshold: number;
    emergency_max_threshold: number;
    humidity_min: number;
    humidity_max: number;
    hysteresis_temp: number;
    hysteresis_hum: number;
}

// Запрос смены режима
export interface ModeRequest {
    mode: 'AUTO' | 'MANUAL';
}

// Системный статус
export interface SystemStatus {
    uptime: number;
    mode: string;
    db_status: string;
}

// Расписание реле
export interface Schedule {
    id: string;
    relay_id: string;
    start_time: string;
    end_time: string;
    is_active: boolean;
    created_at: string;
}

// Запрос создания/обновления расписания
export interface ScheduleRequest {
    relay_id: string;
    start_time: string;
    end_time: string;
    is_active?: boolean;
}

// Отчёт энергопотребления
export interface EnergyReport {
    report_date: string;
    heat_mat_kwh: number;
    light_kwh: number;
    fogger_kwh: number;
    spare_kwh: number;
    total_kwh: number;
}

// Запись журнала переключений реле
export interface RelayLogEntry {
    id: string;
    relay_id: string;
    state: boolean;
    reason: string;
    recorded_at: string;
}

// Стандартная ошибка API
export interface HTTPError {
    code: number;
    message: string;
}

// ID реле для типобезопасных операций
export type RelayId = 'heat_mat' | 'fogger' | 'light' | 'spare';

// Маппинг реле в человекочитаемые названия
export const RELAY_LABELS: Record<RelayId, string> = {
    heat_mat: 'Нагрев',
    fogger: 'Туман',
    light: 'Свет',
    spare: 'Запасной',
};

// Маппинг реле в иконки (emoji)
export const RELAY_ICONS: Record<RelayId, string> = {
    heat_mat: '🔥',
    fogger: '💨',
    light: '💡',
    spare: '🔌',
};
