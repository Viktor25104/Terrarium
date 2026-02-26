package automation

import (
	"context"
	"log"
	"sync"
	"time"

	"terrarium-core/internal/gpio"
	"terrarium-core/internal/models"
	"terrarium-core/internal/storage"
)

// Engine представляет собой ядро, управляющее циклами климат-контроля.
type Engine struct {
	repo       *storage.Repository
	warmSensor gpio.SensorReader
	coldSensor gpio.SensorReader
	heatRelay  gpio.RelayController
	fogRelay   gpio.RelayController
	lightRelay gpio.RelayController

	// mu защищает доступ к кэшированным конфигурациям и показаниям
	mu sync.RWMutex

	// Активный режим: AUTO или MANUAL
	currentMode string

	// Кэш последних показаний датчиков (обновляется каждый цикл)
	lastReadings *models.SensorCurrent
}

// NewEngine инициализирует Конечный Автомат.
func NewEngine(repo *storage.Repository, warmS, coldS gpio.SensorReader, heat, fog, light gpio.RelayController) *Engine {
	return &Engine{
		repo:        repo,
		warmSensor:  warmS,
		coldSensor:  coldS,
		heatRelay:   heat,
		fogRelay:    fog,
		lightRelay:  light,
		currentMode: "AUTO", // По дефолту при старте
	}
}

// GetCurrentReadings возвращает последние показания датчиков из кэша Engine.
// Потокобезопасно — вызывается из HTTP-обработчиков.
func (e *Engine) GetCurrentReadings() *models.SensorCurrent {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.lastReadings
}

// Start запускает фоновую (non-blocking) горутину для контроля климата.
func (e *Engine) Start(ctx context.Context) {
	log.Println("Запуск движка автоматизации климата (Automation Engine)...")

	// Получаем первоначальный режим из БД
	if mode, err := e.repo.GetSystemMode(ctx); err == nil {
		e.currentMode = mode
		log.Printf("[ENGINE] Режим при старте восстановлен: %s\n", e.currentMode)
	}

	// Тикер на опрос датчиков (например, каждые 5 секунд)
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Println("Движок автоматизации остановлен.")
				return
			case <-ticker.C:
				e.evaluateCycle(ctx)
			}
		}
	}()
}

// updateModeCheck синхронизирует режим работы с БД, предотвращая рассинхрон при ручном вызове из API.
func (e *Engine) updateModeCheck(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()
	mode, err := e.repo.GetSystemMode(ctx)
	if err == nil && mode != e.currentMode {
		log.Printf("[ENGINE] Режим изменен %s -> %s\n", e.currentMode, mode)
		e.currentMode = mode
	}
}

// evaluateCycle - одна итерация цикла Конечного Автомата: чтение сенсоров -> проверка безопасности -> гистерезис.
func (e *Engine) evaluateCycle(ctx context.Context) {
	e.updateModeCheck(ctx)

	// ШАГ 1: Чтение датчиков (Сбор данных)
	warmData, errWarm := e.warmSensor.Read()
	coldData, errCold := e.coldSensor.Read()

	if errWarm != nil || errCold != nil {
		log.Println("[ENGINE] ВНИМАНИЕ: Ошибка чтения с датчиков DHT22! Пропускаем цикл.")
		return
	}

	// Обновляем кэш последних показаний (для эндпоинта /sensors/current)
	e.mu.Lock()
	e.lastReadings = &models.SensorCurrent{
		WarmTemp:  warmData.Temperature,
		WarmHum:   warmData.Humidity,
		ColdTemp:  coldData.Temperature,
		ColdHum:   coldData.Humidity,
		Timestamp: time.Now(),
		Mode:      e.currentMode,
	}
	e.mu.Unlock()

	// Пишем лог в базу каждый цикл (5 сек); в проде стоит делать batching
	_ = e.repo.InsertSensorLog(ctx, warmData.Temperature, warmData.Humidity, coldData.Temperature, coldData.Humidity)

	// Читаем текущую конфигурацию (целевые значения) из БД
	cfg, err := e.repo.GetConfig(ctx)
	if err != nil {
		log.Printf("[ENGINE] Невозможно получить конфигурацию из БД: %v. Пропуск цикла.", err)
		return
	}

	// ШАГ 2: БЕЗОПАСНЫЙ (АВАРИЙНЫЙ) КОНТУР - Игнорирует режим (AUTO/MANUAL)! Жизнь важнее.

	// Контур теплового удара
	if warmData.Temperature >= cfg.EmergencyMaxThreshold {
		log.Printf("[EMERGENCY!!!] Температура в теплой зоне %.1f C превысила критическую отметку (%.1f C)!", warmData.Temperature, cfg.EmergencyMaxThreshold)
		_ = e.heatRelay.Off()
		_ = e.fogRelay.Off()
		_ = e.lightRelay.Off() // Свет тоже может греть
		e.repo.InsertRelayLog(ctx, "heat_mat", false, "EMERGENCY_CUTOFF")

		// TODO: Отправить в Telegram Alert
		return // Блокируем дальнейшую логику цикла
	}

	// Контур перегрева холодной зоны (должна оставаться холодной для терморегуляции змеи)
	if coldData.Temperature >= cfg.ColdMaxThreshold {
		log.Printf("[SAFETY] Температура холодной зоны %.1f C превысила предел %.1f C. Отключаем обогрев.", coldData.Temperature, cfg.ColdMaxThreshold)
		if e.heatRelay.IsOn() {
			_ = e.heatRelay.Off()
			e.repo.InsertRelayLog(ctx, "heat_mat", false, "COLD_ZONE_PROTECTION")
		}
	}

	// ШАГ 3: Если режим MANUAL, мы ничего больше не делаем.
	e.mu.RLock()
	mode := e.currentMode
	e.mu.RUnlock()

	if mode == "MANUAL" {
		return
	}

	// ШАГ 4: ЛОГИКА АВТОМАТИЗАЦИИ (РЕЖИМ AUTO - ГИСТЕРЕЗИС)
	e.evaluateHeating(ctx, warmData.Temperature, cfg)
	e.evaluateFogger(ctx, warmData.Humidity, cfg)
}

// evaluateHeating проверяет необходимость включения/выключения термоковрика с учетом гистерезиса
func (e *Engine) evaluateHeating(ctx context.Context, currentTemp float64, cfg *models.ConfigPayload) {
	lowerBound := cfg.WarmTargetMin - cfg.HysteresisTemp
	upperBound := cfg.WarmTargetMax + cfg.HysteresisTemp

	if currentTemp <= lowerBound {
		if !e.heatRelay.IsOn() {
			log.Printf("[AUTO] Температура %.1f упала ниже %.1f. Включаем нагрев.", currentTemp, lowerBound)
			_ = e.heatRelay.On()
			e.repo.InsertRelayLog(ctx, e.heatRelay.Name(), true, "AUTO_TEMP_TRIGGER")
		}
	} else if currentTemp >= upperBound {
		if e.heatRelay.IsOn() {
			log.Printf("[AUTO] Температура %.1f достигла предела %.1f. Отключаем нагрев.", currentTemp, upperBound)
			_ = e.heatRelay.Off()
			e.repo.InsertRelayLog(ctx, e.heatRelay.Name(), false, "AUTO_TEMP_TRIGGER")
		}
	}
}

// evaluateFogger проверяет необходимость вызова дождя/тумана
func (e *Engine) evaluateFogger(ctx context.Context, currentHum float64, cfg *models.ConfigPayload) {
	lowerBound := cfg.HumidityMin - cfg.HysteresisHum
	upperBound := cfg.HumidityMax + cfg.HysteresisHum

	if currentHum <= lowerBound {
		if !e.fogRelay.IsOn() {
			log.Printf("[AUTO] Влажность %.1f%% упала ниже %.1f%%. Включаем генератор тумана.", currentHum, lowerBound)
			_ = e.fogRelay.On()
			e.repo.InsertRelayLog(ctx, e.fogRelay.Name(), true, "AUTO_HUMIDITY_TRIGGER")
		}
	} else if currentHum >= upperBound {
		if e.fogRelay.IsOn() {
			log.Printf("[AUTO] Влажность %.1f%% достигла нормы %.1f%%. Отключаем туман.", currentHum, upperBound)
			_ = e.fogRelay.Off()
			e.repo.InsertRelayLog(ctx, e.fogRelay.Name(), false, "AUTO_HUMIDITY_TRIGGER")
		}
	}
}
