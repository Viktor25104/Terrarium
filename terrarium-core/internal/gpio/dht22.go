package gpio

import (
	"math/rand"
	"time"
)

// SensorData структура для хранения единовременного слепка температуры и влажности.
// Датчики DHT22 отдают эти параметры одновременно, поэтому мы опрашиваем пин один раз
// и сразу заполняем оба поля.
type SensorData struct {
	Temperature float64 // в градусах Цельсия
	Humidity    float64 // в процентах (0-100)
	Timestamp   time.Time
}

// SensorReader описывает интерфейс любого температурного датчика.
type SensorReader interface {
	// Read берет данные с сенсора. Может вернуть ошибку, если сенсор отключен.
	Read() (SensorData, error)
	// Name возвращает имя датчика (например "WarmZone" или "ColdZone")
	Name() string
}

// ==========================================
// MOCK РЕАЛИЗАЦИЯ (для ПК / Windows / Mac)
// ==========================================

// MockDHT22 имитирует работу реального датчика.
type MockDHT22 struct {
	sensorName string
	baseTemp   float64
	baseHum    float64
}

// NewMockDHT22 создает новый мок-датчик (идеально для локальной разработки без Raspberry).
func NewMockDHT22(name string, baseTemp, baseHum float64) *MockDHT22 {
	// Инициализация генератора рандома для фейковых колебаний
	rand.Seed(time.Now().UnixNano())
	return &MockDHT22{
		sensorName: name,
		baseTemp:   baseTemp,
		baseHum:    baseHum,
	}
}

func (m *MockDHT22) Name() string { return m.sensorName }

func (m *MockDHT22) Read() (SensorData, error) {
	// Генерация легких колебаний (от -0.5 до +0.5 градуса/процента)
	tempJitter := (rand.Float64() * 1.0) - 0.5
	humJitter := (rand.Float64() * 2.0) - 1.0

	// Имитируем "медленное чтение" физического сенсора
	time.Sleep(100 * time.Millisecond)

	return SensorData{
		Temperature: m.baseTemp + tempJitter,
		Humidity:    m.baseHum + humJitter,
		Timestamp:   time.Now(),
	}, nil
}

// ==========================================
// РЕАЛИЗАЦИЯ ДЛЯ LIBGPIOD ОСТАЕТСЯ ПОКА В TODO
// ==========================================
// Для реального железа потребуется вызов CGO (libgpiod), который будет
// реализован в структуре `RealDHT22`. Пока что используем MockDHT22
// для сборки и написания логики управления.
