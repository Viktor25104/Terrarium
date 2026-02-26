package models

import "time"

// HTTPError представляет стандартную структуру ошибки API.
// Используется для возврата детальной информации в случае проблем (например, 400 Bad Request или 500 Internal Server Error).
// @Description Стандартный формат возвращаемой ошибки при нештатных или невалидных запросах
type HTTPError struct {
	// HTTP код статуса ошибки
	// Example: 400
	Code int `json:"code" example:"400"`
	// Текстовое описание ошибки или причина отказа
	// Example: "Invalid request payload"
	Message string `json:"message" example:"Параметры выходят за допустимые пределы"`
}

// ConfigPayload представляет набор настроек и порогов автоматизации для террариума.
// Описывает температурные и влажностные режимы, а также границы аварийных отключений.
// @Description Payload конфигурации для управления поведением механизма климат-контроля
type ConfigPayload struct {
	// Минимальная целевая температура в теплой зоне (°C), при которой включается обогрев.
	// Ограничения: от 20.0 до 40.0.
	// Example: 31.5
	WarmTargetMin float64 `json:"warm_target_min" binding:"required,min=20,max=40" example:"31.5"`
	// Максимальная целевая температура в теплой зоне (°C), при достижении которой обогрев отключается.
	// Ограничения: от WarmTargetMin до 40.0.
	// Example: 33.0
	WarmTargetMax float64 `json:"warm_target_max" binding:"required,min=20,max=40" example:"33.0"`
	// Максимально допустимая температура в холодной зоне (°C). Если превышена, обогрев принудительно отключается.
	// Example: 26.5
	ColdMaxThreshold float64 `json:"cold_max_threshold" binding:"required,min=20,max=35" example:"26.5"`
	// Температурный порог теплой зоны (°C), при котором система аварийно отключает всё и шлёт Alert в Telegram.
	// Example: 35.0
	EmergencyMaxThreshold float64 `json:"emergency_max_threshold" binding:"required,min=30,max=45" example:"35.0"`
	// Минимальная влажность (%), при которой включается фоггер (генератор тумана).
	// Example: 50.0
	HumidityMin float64 `json:"humidity_min" binding:"required,min=0,max=100" example:"50.0"`
	// Максимальная влажность (%), при достижении которой фоггер отключается.
	// Example: 65.0
	HumidityMax float64 `json:"humidity_max" binding:"required,min=0,max=100" example:"65.0"`
	// Температурный гистерезис (°C), чтобы избежать "дребезга" реле около целевого значения.
	// Example: 0.5
	HysteresisTemp float64 `json:"hysteresis_temp" binding:"required,min=0.1,max=5" example:"0.5"`
	// Гистерезис влажности (%), для предотвращения частого срабатывания фоггера.
	// Example: 2.0
	HysteresisHum float64 `json:"hysteresis_hum" binding:"required,min=0.5,max=10" example:"2.0"`
}

// ModeRequest представляет запрос на переключение режима работы террариума.
// @Description Запрос для переключения между АВТОМАТИЧЕСКОЙ и РУЧНОЙ работой механизмов.
type ModeRequest struct {
	// Целевой режим работы системы. Допускается 'AUTO' (автоматика) или 'MANUAL' (ручное управление).
	// Example: MANUAL
	Mode string `json:"mode" binding:"required,oneof=AUTO MANUAL" example:"MANUAL"`
}

// SystemStatus представляет текущее операционное состояние Backend'а.
// @Description Состояние системы, режим и аптайм
type SystemStatus struct {
	// Время работы сервиса с момента старта (в секундах).
	// Example: 3600
	Uptime int64 `json:"uptime" example:"3600"`
	// Текущий активный режим автоматизации (AUTO или MANUAL).
	// Example: AUTO
	Mode string `json:"mode" example:"AUTO"`
	// Статус соединения с базой данных PostgreSQL (OK или ERROR).
	// Example: OK
	DBStatus string `json:"db_status" example:"OK"`
}

// RelayState описывает текущее состояние аппаратных реле, подключенных к Raspberry Pi.
// @Description Фактическое (электрическое) состояние исполнительных устройств.
type RelayState struct {
	// Состояние реле термоковрика (true = ВКЛ, false = ВЫКЛ)
	// Example: true
	HeatMat bool `json:"heat_mat" example:"true"`
	// Состояние реле генератора тумана
	// Example: false
	Fogger bool `json:"fogger" example:"false"`
	// Состояние освещения террариума
	// Example: true
	Light bool `json:"light" example:"true"`
	// Состояние запасной розетки
	// Example: false
	Spare bool `json:"spare" example:"false"`
}

// RelayToggleRequest представляет запрос пользователя на изменение состояния конкретного реле в ручном режиме.
// @Description Запрос на принудительное переключение состояния реле
type RelayToggleRequest struct {
	// Желаемое состояние реле (true для включения, false для выключения)
	// Example: true
	State bool `json:"state" example:"true"`
}

// SensorDataHistory содержит агрегированные данные датчиков за указанный период времени.
// @Description Историческая справка температуры и влажности из БД для графиков.
type SensorDataHistory struct {
	// Точное время и дата, когда показания были записаны системой
	// Example: "2026-02-26T13:30:00Z"
	Timestamp time.Time `json:"timestamp" example:"2026-02-26T13:30:00Z"`
	// Температура (°C) с теплой зоны террариума
	// Example: 32.1
	WarmTemp float64 `json:"warm_temp" example:"32.1"`
	// Влажность (%) с теплой зоны
	// Example: 55.4
	WarmHum float64 `json:"warm_hum" example:"55.4"`
	// Температура (°C) с холодной зоны террариума (для безопасности)
	// Example: 25.0
	ColdTemp float64 `json:"cold_temp" example:"25.0"`
	// Влажность (%) с холодной зоны
	// Example: 60.1
	ColdHum float64 `json:"cold_hum" example:"60.1"`
}

// EnergyReport представляет агрегированный список энергопотребления по каждому компоненту.
// @Description Общие затраты энергопотребления террариумом (рассчитываются из времени работы и заявленной мощности реле).
type EnergyReport struct {
	// Дата генерации отчета
	// Example: "2026-02-26"
	Date string `json:"report_date" example:"2026-02-26"`
	// КВт⋅ч, затраченные нагревателем (Heat Mat)
	// Example: 0.1500
	HeatMatKwh float64 `json:"heat_mat_kwh" example:"0.1500"`
	// КВт⋅ч, затраченные освещением
	// Example: 0.2400
	LightKwh float64 `json:"light_kwh" example:"0.2400"`
	// КВт⋅ч, затраченные генератором тумана
	// Example: 0.0300
	FoggerKwh float64 `json:"fogger_kwh" example:"0.0300"`
	// КВт⋅ч, затраченные запасным портом
	// Example: 0.0000
	SpareKwh float64 `json:"spare_kwh" example:"0.0000"`
	// Суммарные затраты по всем розеткам (КВт⋅ч)
	// Example: 0.4200
	TotalKwh float64 `json:"total_kwh" example:"0.4200"`
}

// SensorCurrent представляет актуальные (последние) показания с обоих датчиков DHT22.
// Каждый датчик измеряет температуру и влажность одновременно.
// @Description Текущие (live) показания температуры и влажности с двух зон террариума.
type SensorCurrent struct {
	// Температура (°C) в тёплой зоне
	// Example: 32.3
	WarmTemp float64 `json:"warm_temp" example:"32.3"`
	// Влажность (%) в тёплой зоне
	// Example: 58.5
	WarmHum float64 `json:"warm_hum" example:"58.5"`
	// Температура (°C) в холодной зоне
	// Example: 24.8
	ColdTemp float64 `json:"cold_temp" example:"24.8"`
	// Влажность (%) в холодной зоне
	// Example: 65.2
	ColdHum float64 `json:"cold_hum" example:"65.2"`
	// Время последнего считывания с датчиков
	// Example: "2026-02-26T15:30:00Z"
	Timestamp time.Time `json:"timestamp" example:"2026-02-26T15:30:00Z"`
	// Текущий режим системы (AUTO / MANUAL)
	// Example: AUTO
	Mode string `json:"mode" example:"AUTO"`
}

// Schedule представляет запись расписания включения/выключения реле (например, освещение по таймеру).
// @Description Расписание автоматического управления реле по времени суток.
type Schedule struct {
	// Уникальный идентификатор расписания (UUID)
	// Example: "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	ID string `json:"id" example:"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
	// ID реле, к которому привязано расписание
	// Example: "light"
	RelayID string `json:"relay_id" example:"light"`
	// Время включения (формат HH:MM)
	// Example: "08:00"
	StartTime string `json:"start_time" example:"08:00"`
	// Время выключения (формат HH:MM)
	// Example: "20:00"
	EndTime string `json:"end_time" example:"20:00"`
	// Активно ли расписание
	// Example: true
	IsActive bool `json:"is_active" example:"true"`
	// Дата создания записи
	// Example: "2026-02-26T12:00:00Z"
	CreatedAt time.Time `json:"created_at" example:"2026-02-26T12:00:00Z"`
}

// ScheduleRequest представляет запрос на создание или обновление расписания реле.
// @Description Payload для создания/обновления расписания реле.
type ScheduleRequest struct {
	// ID реле (heat_mat, fogger, light, spare)
	// Example: "light"
	RelayID string `json:"relay_id" binding:"required" example:"light"`
	// Время включения (формат HH:MM)
	// Example: "08:00"
	StartTime string `json:"start_time" binding:"required" example:"08:00"`
	// Время выключения (формат HH:MM)
	// Example: "20:00"
	EndTime string `json:"end_time" binding:"required" example:"20:00"`
	// Активно ли расписание (по умолчанию true)
	// Example: true
	IsActive *bool `json:"is_active" example:"true"`
}

// RelayLogEntry представляет одну запись аудита включения/выключения реле.
// @Description Запись журнала переключений реле с причиной и временной меткой.
type RelayLogEntry struct {
	// Уникальный идентификатор записи (UUID)
	// Example: "f1e2d3c4-b5a6-7890-cdef-1234567890ab"
	ID string `json:"id" example:"f1e2d3c4-b5a6-7890-cdef-1234567890ab"`
	// Идентификатор реле (heat_mat / fogger / light / spare)
	// Example: "heat_mat"
	RelayID string `json:"relay_id" example:"heat_mat"`
	// Состояние реле после переключения (true = ВКЛ, false = ВЫКЛ)
	// Example: true
	State bool `json:"state" example:"true"`
	// Причина переключения (AUTO_TEMP_TRIGGER, MANUAL_OVERRIDE, EMERGENCY_CUTOFF и т.д.)
	// Example: "AUTO_TEMP_TRIGGER"
	Reason string `json:"reason" example:"AUTO_TEMP_TRIGGER"`
	// Время события
	// Example: "2026-02-26T14:05:00Z"
	RecordedAt time.Time `json:"recorded_at" example:"2026-02-26T14:05:00Z"`
}
