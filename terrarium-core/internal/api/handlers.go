package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"terrarium-core/internal/automation"
	"terrarium-core/internal/gpio"
	"terrarium-core/internal/models"
	"terrarium-core/internal/storage"

	"github.com/gin-gonic/gin"
)

// API struct содержит все зависимости (БД, Реле и Движок), необходимые для обработки HTTP-запросов.
type API struct {
	Repo   *storage.Repository
	Relays map[string]gpio.RelayController
	Engine *automation.Engine
}

// ==========================================
// SYSTEM & CONFIGURATION
// ==========================================

// GetConfig godoc
// @Summary Получить текущую конфигурацию (Настройки Автоматизации)
// @Description Возвращает текущие настройки террариума: полярные целевые значения температуры, влажности, гистерезиса и пороги аварийных отключений. Настройки подтягиваются из Postgres.
// @Tags System, Configuration
// @Accept json
// @Produce json
// @Success 200 {object} models.ConfigPayload "Успешное получение настроек"
// @Failure 500 {object} models.HTTPError "Внутренняя ошибка (например, сбой соединения с БД)"
// @Router /api/v1/config [get]
func (a *API) GetConfig(c *gin.Context) {
	cfg, err := a.Repo.GetConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка чтения БД"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// UpdateConfig godoc
// @Summary Обновить границы климатического контроля
// @Description Принимает новые пороговые значения (Payload) и валидирует их. В случае успеха, новые пороги сохраняются в БД.
// @Tags System, Configuration
// @Accept json
// @Produce json
// @Param payload body models.ConfigPayload true "Объект новых настроек климата"
// @Success 200 {object} models.ConfigPayload "Конфигурация успешно обновлена"
// @Failure 400 {object} models.HTTPError "Невалидный Payload"
// @Failure 500 {object} models.HTTPError "Ошибка записи в базу данных postgres"
// @Router /api/v1/config [put]
func (a *API) UpdateConfig(c *gin.Context) {
	var cfg models.ConfigPayload
	if err := c.ShouldBindJSON(&cfg); err != nil {
		c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: err.Error()})
		return
	}

	// Бизнес-логика валидации
	if cfg.WarmTargetMax <= cfg.WarmTargetMin {
		c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: "WarmTargetMax должен быть больше WarmTargetMin"})
		return
	}

	if err := a.Repo.UpdateConfig(c.Request.Context(), cfg); err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка записи в БД"})
		return
	}

	c.JSON(http.StatusOK, cfg)
}

// GetSystemStatus godoc
// @Summary Получить статус и общую "проверку здоровья" (Health check) системы
// @Description Предоставляет uptime приложения и текущий режим работы автомата (AUTO/MANUAL) из БД.
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} models.SystemStatus "Системный статус успешно получен"
// @Router /api/v1/system/status [get]
func (a *API) GetSystemStatus(c *gin.Context) {
	mode, err := a.Repo.GetSystemMode(c.Request.Context())
	dbStat := "OK"
	if err != nil {
		dbStat = "ERROR"
		mode = "UNKNOWN"
	}

	status := models.SystemStatus{
		Uptime:   999, // TODO: Реализовать глобальный счетчик Uptime
		Mode:     mode,
		DBStatus: dbStat,
	}
	c.JSON(http.StatusOK, status)
}

// SetSystemMode godoc
// @Summary Изменить глобальный режим системы (AUTO или MANUAL)
// @Description Позволяет пользователю полностью перехватить контроль над реле.
// @Tags System
// @Accept json
// @Produce json
// @Param payload body models.ModeRequest true "Целевой режим: 'AUTO' или 'MANUAL'"
// @Success 200 {object} models.ModeRequest "Режим успешно изменен"
// @Failure 400 {object} models.HTTPError "Неверно заданный режим"
// @Router /api/v1/system/mode [post]
func (a *API) SetSystemMode(c *gin.Context) {
	var req models.ModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: err.Error()})
		return
	}

	if err := a.Repo.SetSystemMode(c.Request.Context(), req.Mode); err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка сохранения режима"})
		return
	}
	c.JSON(http.StatusOK, req)
}

// ==========================================
// RELAYS (HARDWARE CONTROL)
// ==========================================

// GetRelays godoc
// @Summary Получить фактическое состояние пинов GPIO (всех 4 реле)
// @Description Выполняет прямой опрос состояния реле из памяти/оборудования.
// @Tags Hardware Control (Manual Mode)
// @Accept json
// @Produce json
// @Success 200 {object} models.RelayState "Состояние реле"
// @Router /api/v1/relays [get]
func (a *API) GetRelays(c *gin.Context) {
	state := models.RelayState{
		HeatMat: a.Relays["heat_mat"].IsOn(),
		Fogger:  a.Relays["fogger"].IsOn(),
		Light:   a.Relays["light"].IsOn(),
		Spare:   a.Relays["spare"].IsOn(),
	}
	c.JSON(http.StatusOK, state)
}

// ToggleRelay godoc
// @Summary Переключить конкретное реле [Требует MANUAL режим]
// @Description Сигнализирует Raspberry Pi переключить уровень GPIO на конкретном пине. Работает только в MANUAL.
// @Tags Hardware Control (Manual Mode)
// @Accept json
// @Produce json
// @Param id path string true "ID реле для переключения" Enums(heat_mat, fogger, light, spare)
// @Param payload body models.RelayToggleRequest true "Телеметрия с запросом на активацию/деактивацию"
// @Success 200 {string} string "Реле успешно переключено"
// @Failure 400 {object} models.HTTPError "Неизвестный ID реле"
// @Failure 403 {object} models.HTTPError "Система находится в режиме AUTO (ручное управление запрещено)"
// @Router /api/v1/relays/{id}/toggle [post]
func (a *API) ToggleRelay(c *gin.Context) {
	relayID := c.Param("id")

	// Проверка режима (если AUTO -> вернуть 403)
	mode, err := a.Repo.GetSystemMode(c.Request.Context())
	if err != nil || mode != "MANUAL" {
		c.JSON(http.StatusForbidden, models.HTTPError{Code: 403, Message: "Ручное переключение разрешено только в режиме MANUAL"})
		return
	}

	var req models.RelayToggleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: err.Error()})
		return
	}

	relay, exists := a.Relays[relayID]
	if !exists {
		c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: "Неизвестное реле: " + relayID})
		return
	}

	if req.State {
		_ = relay.On()
	} else {
		_ = relay.Off()
	}

	// Запись лога переключения
	_ = a.Repo.InsertRelayLog(c.Request.Context(), relayID, req.State, "MANUAL_OVERRIDE")

	c.JSON(http.StatusOK, gin.H{"msg": fmt.Sprintf("Реле %s переведено в %t", relayID, req.State)})
}

// ==========================================
// SENSORS (ТЕКУЩИЕ ПОКАЗАНИЯ + ИСТОРИЯ)
// ==========================================

// GetSensorCurrent godoc
// @Summary Получить текущие показания датчиков (температура + влажность, обе зоны)
// @Description Возвращает последние мгновенные показания с DHT22. Данные кэшируются в Engine (обновляются каждые 5 сек). При отсутствии подключения к оборудованию возвращаются mock-значения.
// @Tags Sensors
// @Produce json
// @Success 200 {object} models.SensorCurrent "Текущие показания"
// @Router /api/v1/sensors/current [get]
func (a *API) GetSensorCurrent(c *gin.Context) {
	readings := a.Engine.GetCurrentReadings()
	if readings == nil {
		// Движок ещё не успел сделать первый цикл — возвращаем пустые данные
		c.JSON(http.StatusOK, models.SensorCurrent{
			Timestamp: time.Now(),
			Mode:      "UNKNOWN",
		})
		return
	}
	c.JSON(http.StatusOK, readings)
}

// GetSensorMetrics godoc
// @Summary Получить историю показаний датчиков
// @Description Возвращает исторические данные температуры и влажности из sensor_logs. Поддерживает фильтрацию по дате и ограничение выборки. Если система подключена — данные реальные из БД; если нет — массив будет пуст.
// @Tags Metrics
// @Produce json
// @Param from query string false "Начало периода (RFC3339, например 2026-02-26T00:00:00Z)"
// @Param to query string false "Конец периода (RFC3339, например 2026-02-26T23:59:59Z)"
// @Param limit query int false "Максимальное количество записей (по умолчанию 100, макс 1000)"
// @Success 200 {array} models.SensorDataHistory "Историческая справка"
// @Failure 400 {object} models.HTTPError "Неверный формат параметров"
// @Failure 500 {object} models.HTTPError "Ошибка чтения из БД"
// @Router /api/v1/metrics/sensors [get]
func (a *API) GetSensorMetrics(c *gin.Context) {
	var from, to time.Time
	var err error

	if fromStr := c.Query("from"); fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: "Неверный формат 'from': " + err.Error()})
			return
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: "Неверный формат 'to': " + err.Error()})
			return
		}
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}

	data, err := a.Repo.GetSensorHistory(c.Request.Context(), from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка чтения метрик: " + err.Error()})
		return
	}

	// Гарантируем пустой массив вместо null в JSON
	if data == nil {
		data = []models.SensorDataHistory{}
	}
	c.JSON(http.StatusOK, data)
}

// ==========================================
// ENERGY (ЭНЕРГОПОТРЕБЛЕНИЕ)
// ==========================================

// GetEnergyMetrics godoc
// @Summary Получить отчёты энергопотребления
// @Description Возвращает агрегированные отчёты расхода электроэнергии по каждому реле (кВт⋅ч). Данные берутся из таблицы energy_reports. Если отчёты ещё не генерировались — массив будет пуст.
// @Tags Metrics
// @Produce json
// @Param from query string false "Начало периода (формат YYYY-MM-DD)"
// @Param to query string false "Конец периода (формат YYYY-MM-DD)"
// @Success 200 {array} models.EnergyReport "Отчёты энергопотребления"
// @Failure 500 {object} models.HTTPError "Ошибка чтения из БД"
// @Router /api/v1/metrics/energy [get]
func (a *API) GetEnergyMetrics(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	data, err := a.Repo.GetEnergyReports(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка чтения отчётов: " + err.Error()})
		return
	}

	if data == nil {
		data = []models.EnergyReport{}
	}
	c.JSON(http.StatusOK, data)
}

// ==========================================
// SCHEDULES (РАСПИСАНИЯ РЕЛЕ)
// ==========================================

// GetSchedules godoc
// @Summary Получить все расписания реле
// @Description Возвращает список всех расписаний автоматического включения/выключения реле по времени.
// @Tags Schedules
// @Produce json
// @Success 200 {array} models.Schedule "Список расписаний"
// @Failure 500 {object} models.HTTPError "Ошибка чтения из БД"
// @Router /api/v1/schedules [get]
func (a *API) GetSchedules(c *gin.Context) {
	schedules, err := a.Repo.GetSchedules(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка чтения расписаний: " + err.Error()})
		return
	}

	if schedules == nil {
		schedules = []models.Schedule{}
	}
	c.JSON(http.StatusOK, schedules)
}

// CreateSchedule godoc
// @Summary Создать новое расписание реле
// @Description Создаёт запись расписания для автоматического включения/выключения реле по времени суток.
// @Tags Schedules
// @Accept json
// @Produce json
// @Param payload body models.ScheduleRequest true "Данные расписания"
// @Success 201 {object} models.Schedule "Расписание успешно создано"
// @Failure 400 {object} models.HTTPError "Невалидный Payload"
// @Failure 500 {object} models.HTTPError "Ошибка записи в БД"
// @Router /api/v1/schedules [post]
func (a *API) CreateSchedule(c *gin.Context) {
	var req models.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: err.Error()})
		return
	}

	schedule, err := a.Repo.CreateSchedule(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка создания расписания: " + err.Error()})
		return
	}
	c.JSON(http.StatusCreated, schedule)
}

// UpdateSchedule godoc
// @Summary Обновить существующее расписание
// @Description Изменяет параметры существующего расписания реле по его UUID.
// @Tags Schedules
// @Accept json
// @Produce json
// @Param id path string true "UUID расписания"
// @Param payload body models.ScheduleRequest true "Обновлённые данные"
// @Success 200 {string} string "Расписание обновлено"
// @Failure 400 {object} models.HTTPError "Невалидный Payload"
// @Failure 404 {object} models.HTTPError "Расписание не найдено"
// @Failure 500 {object} models.HTTPError "Ошибка обновления"
// @Router /api/v1/schedules/{id} [put]
func (a *API) UpdateSchedule(c *gin.Context) {
	id := c.Param("id")

	var req models.ScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.HTTPError{Code: 400, Message: err.Error()})
		return
	}

	if err := a.Repo.UpdateSchedule(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusNotFound, models.HTTPError{Code: 404, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Расписание обновлено"})
}

// DeleteSchedule godoc
// @Summary Удалить расписание реле
// @Description Удаляет запись расписания по UUID.
// @Tags Schedules
// @Produce json
// @Param id path string true "UUID расписания"
// @Success 200 {string} string "Расписание удалено"
// @Failure 404 {object} models.HTTPError "Расписание не найдено"
// @Router /api/v1/schedules/{id} [delete]
func (a *API) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")

	if err := a.Repo.DeleteSchedule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, models.HTTPError{Code: 404, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "Расписание удалено"})
}

// ==========================================
// RELAY LOGS (ЖУРНАЛ ПЕРЕКЛЮЧЕНИЙ)
// ==========================================

// GetRelayLogs godoc
// @Summary Получить журнал переключений реле
// @Description Возвращает аудит-лог всех событий включения/выключения реле с причиной и временной меткой. Поддерживает пагинацию.
// @Tags Logs
// @Produce json
// @Param limit query int false "Количество записей (по умолчанию 50, макс 500)"
// @Param offset query int false "Смещение для пагинации (по умолчанию 0)"
// @Success 200 {array} models.RelayLogEntry "Журнал переключений"
// @Failure 500 {object} models.HTTPError "Ошибка чтения из БД"
// @Router /api/v1/relay-logs [get]
func (a *API) GetRelayLogs(c *gin.Context) {
	limit := 50
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil {
			limit = parsed
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsed
		}
	}

	logs, err := a.Repo.GetRelayLogs(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.HTTPError{Code: 500, Message: "Ошибка чтения логов: " + err.Error()})
		return
	}

	if logs == nil {
		logs = []models.RelayLogEntry{}
	}
	c.JSON(http.StatusOK, logs)
}
