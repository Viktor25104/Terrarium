package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"terrarium-core/internal/models"
)

// Repository обеспечивает слой абстракции над SQL-запросами к PostgreSQL
type Repository struct {
	db *DB
}

// NewRepository создает новый инстанс репозитория
func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

// GetConfig извлекает единственную активную конфигурацию климата из БД.
func (r *Repository) GetConfig(ctx context.Context) (*models.ConfigPayload, error) {
	query := `
		SELECT 
			warm_target_min, warm_target_max, cold_max_threshold, emergency_max_threshold,
			humidity_min, humidity_max, hysteresis_temp, hysteresis_hum
		FROM automation_settings 
		WHERE id = 1
	`
	var cfg models.ConfigPayload
	err := r.db.Pool.QueryRow(ctx, query).Scan(
		&cfg.WarmTargetMin, &cfg.WarmTargetMax, &cfg.ColdMaxThreshold, &cfg.EmergencyMaxThreshold,
		&cfg.HumidityMin, &cfg.HumidityMax, &cfg.HysteresisTemp, &cfg.HysteresisHum,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения конфигурации из БД: %w", err)
	}
	return &cfg, nil
}

// UpdateConfig обновляет текущую конфигурацию климата.
func (r *Repository) UpdateConfig(ctx context.Context, cfg models.ConfigPayload) error {
	query := `
		UPDATE automation_settings 
		SET 
			warm_target_min = $1, warm_target_max = $2, 
			cold_max_threshold = $3, emergency_max_threshold = $4,
			humidity_min = $5, humidity_max = $6, 
			hysteresis_temp = $7, hysteresis_hum = $8,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`
	_, err := r.db.Pool.Exec(ctx, query,
		cfg.WarmTargetMin, cfg.WarmTargetMax,
		cfg.ColdMaxThreshold, cfg.EmergencyMaxThreshold,
		cfg.HumidityMin, cfg.HumidityMax,
		cfg.HysteresisTemp, cfg.HysteresisHum,
	)
	if err != nil {
		return fmt.Errorf("ошибка обновления конфигурации: %w", err)
	}
	return nil
}

// GetSystemMode возвращает текущий режим работы автоматики (AUTO / MANUAL).
func (r *Repository) GetSystemMode(ctx context.Context) (string, error) {
	var mode string
	err := r.db.Pool.QueryRow(ctx, `SELECT mode FROM automation_settings WHERE id = 1`).Scan(&mode)
	if err != nil {
		return "AUTO", err // По умолчанию всегда AUTO в случае сбоя чтения
	}
	return mode, nil
}

// SetSystemMode переключает режим работы между AUTO и MANUAL.
func (r *Repository) SetSystemMode(ctx context.Context, mode string) error {
	_, err := r.db.Pool.Exec(ctx, `UPDATE automation_settings SET mode = $1, updated_at = CURRENT_TIMESTAMP WHERE id = 1`, mode)
	return err
}

// InsertSensorLog сохраняет показания обоих датчиков в Timeseries таблицу.
func (r *Repository) InsertSensorLog(ctx context.Context, warmTemp, warmHum, coldTemp, coldHum float64) error {
	query := `
		INSERT INTO sensor_logs (warm_zone_temp, warm_zone_hum, cold_zone_temp, cold_zone_hum)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Pool.Exec(ctx, query, warmTemp, warmHum, coldTemp, coldHum)
	if err != nil {
		log.Printf("Ошибка сохранения лога датчиков: %v\n", err)
	}
	return err
}

// InsertRelayLog записывает в аудит событие переключения релейного аппарата.
func (r *Repository) InsertRelayLog(ctx context.Context, relayID string, state bool, reason string) error {
	query := `
		INSERT INTO relay_logs (relay_id, state, reason)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Pool.Exec(ctx, query, relayID, state, reason)
	if err != nil {
		log.Printf("Ошибка сохранения лога реле: %v\n", err)
	}
	return err
}

// GetSensorHistory возвращает исторические показания датчиков за указанный период.
// Если from/to не заданы (zero), возвращает последние записи с учётом limit.
func (r *Repository) GetSensorHistory(ctx context.Context, from, to time.Time, limit int) ([]models.SensorDataHistory, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}

	var query string
	var args []interface{}

	if !from.IsZero() && !to.IsZero() {
		query = `
			SELECT recorded_at, warm_zone_temp, warm_zone_hum, cold_zone_temp, cold_zone_hum
			FROM sensor_logs
			WHERE recorded_at BETWEEN $1 AND $2
			ORDER BY recorded_at DESC
			LIMIT $3
		`
		args = []interface{}{from, to, limit}
	} else {
		query = `
			SELECT recorded_at, warm_zone_temp, warm_zone_hum, cold_zone_temp, cold_zone_hum
			FROM sensor_logs
			ORDER BY recorded_at DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выборки истории датчиков: %w", err)
	}
	defer rows.Close()

	var result []models.SensorDataHistory
	for rows.Next() {
		var entry models.SensorDataHistory
		if err := rows.Scan(&entry.Timestamp, &entry.WarmTemp, &entry.WarmHum, &entry.ColdTemp, &entry.ColdHum); err != nil {
			return nil, fmt.Errorf("ошибка чтения строки sensor_logs: %w", err)
		}
		result = append(result, entry)
	}
	return result, nil
}

// GetEnergyReports возвращает агрегированные отчёты энергопотребления за указанный период.
func (r *Repository) GetEnergyReports(ctx context.Context, from, to string) ([]models.EnergyReport, error) {
	var query string
	var args []interface{}

	if from != "" && to != "" {
		query = `
			SELECT report_date, heat_mat_kwh, light_kwh, fogger_kwh, spare_kwh, total_kwh
			FROM energy_reports
			WHERE report_date BETWEEN $1 AND $2
			ORDER BY report_date DESC
		`
		args = []interface{}{from, to}
	} else {
		query = `
			SELECT report_date, heat_mat_kwh, light_kwh, fogger_kwh, spare_kwh, total_kwh
			FROM energy_reports
			ORDER BY report_date DESC
			LIMIT 30
		`
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("ошибка выборки отчётов энергопотребления: %w", err)
	}
	defer rows.Close()

	var result []models.EnergyReport
	for rows.Next() {
		var entry models.EnergyReport
		if err := rows.Scan(&entry.Date, &entry.HeatMatKwh, &entry.LightKwh, &entry.FoggerKwh, &entry.SpareKwh, &entry.TotalKwh); err != nil {
			return nil, fmt.Errorf("ошибка чтения строки energy_reports: %w", err)
		}
		result = append(result, entry)
	}
	return result, nil
}

// GetSchedules возвращает все расписания реле.
func (r *Repository) GetSchedules(ctx context.Context) ([]models.Schedule, error) {
	query := `
		SELECT id, relay_id, start_time, end_time, is_active, created_at
		FROM schedules
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка выборки расписаний: %w", err)
	}
	defer rows.Close()

	var result []models.Schedule
	for rows.Next() {
		var s models.Schedule
		var startTime, endTime time.Time
		if err := rows.Scan(&s.ID, &s.RelayID, &startTime, &endTime, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка чтения расписания: %w", err)
		}
		s.StartTime = startTime.Format("15:04")
		s.EndTime = endTime.Format("15:04")
		result = append(result, s)
	}
	return result, nil
}

// CreateSchedule создаёт новое расписание реле и возвращает созданную запись.
func (r *Repository) CreateSchedule(ctx context.Context, req models.ScheduleRequest) (*models.Schedule, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	query := `
		INSERT INTO schedules (relay_id, start_time, end_time, is_active)
		VALUES ($1, $2::time, $3::time, $4)
		RETURNING id, relay_id, start_time, end_time, is_active, created_at
	`
	var s models.Schedule
	var startTime, endTime time.Time
	err := r.db.Pool.QueryRow(ctx, query, req.RelayID, req.StartTime, req.EndTime, isActive).
		Scan(&s.ID, &s.RelayID, &startTime, &endTime, &s.IsActive, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания расписания: %w", err)
	}
	s.StartTime = startTime.Format("15:04")
	s.EndTime = endTime.Format("15:04")
	return &s, nil
}

// UpdateSchedule обновляет существующее расписание по ID.
func (r *Repository) UpdateSchedule(ctx context.Context, id string, req models.ScheduleRequest) error {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	query := `
		UPDATE schedules
		SET relay_id = $1, start_time = $2::time, end_time = $3::time, is_active = $4
		WHERE id = $5
	`
	ct, err := r.db.Pool.Exec(ctx, query, req.RelayID, req.StartTime, req.EndTime, isActive, id)
	if err != nil {
		return fmt.Errorf("ошибка обновления расписания: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("расписание с id=%s не найдено", id)
	}
	return nil
}

// DeleteSchedule удаляет расписание по ID.
func (r *Repository) DeleteSchedule(ctx context.Context, id string) error {
	ct, err := r.db.Pool.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("ошибка удаления расписания: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("расписание с id=%s не найдено", id)
	}
	return nil
}

// GetRelayLogs возвращает журнал переключений реле с пагинацией.
func (r *Repository) GetRelayLogs(ctx context.Context, limit, offset int) ([]models.RelayLogEntry, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, relay_id, state, reason, recorded_at
		FROM relay_logs
		ORDER BY recorded_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка выборки логов реле: %w", err)
	}
	defer rows.Close()

	var result []models.RelayLogEntry
	for rows.Next() {
		var entry models.RelayLogEntry
		if err := rows.Scan(&entry.ID, &entry.RelayID, &entry.State, &entry.Reason, &entry.RecordedAt); err != nil {
			return nil, fmt.Errorf("ошибка чтения строки relay_logs: %w", err)
		}
		result = append(result, entry)
	}
	return result, nil
}
