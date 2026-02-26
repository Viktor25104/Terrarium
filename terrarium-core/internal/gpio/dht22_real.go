package gpio

import (
	"fmt"
	"time"

	"github.com/MichaelS11/go-dht"
)

// RealDHT22 обеспечивает чтение температуры и влажности с реального диода по 1-wire
type RealDHT22 struct {
	name   string
	pinIdx int
}

// NewRealDHT22 инициализирует аппаратный датчик.
func NewRealDHT22(name string, pinOffset int) (*RealDHT22, error) {
	// Под капотом библиотека go-dht требует инициализации host (использует /dev/mem или sysfs)
	// В современных Pi 5 лучше использовать gpiod, но протокол протокол DHT очень требователен к таймингам,
	// и специализированная либа справляется лучше всего.
	err := dht.HostInit()
	if err != nil {
		return nil, fmt.Errorf("ошибка HostInit() для датчиков: %w", err)
	}

	return &RealDHT22{
		name:   name,
		pinIdx: pinOffset,
	}, nil
}

func (d *RealDHT22) Name() string {
	return d.name
}

// Read опрашивает датчик. Функция блокирующая, занимает ~1-2 секунды,
// так как датчик медленно отдает биты данных.
func (d *RealDHT22) Read() (SensorData, error) {
	// Используем номер GPIO (по BCM)
	pinName := fmt.Sprintf("GPIO%d", d.pinIdx)

	// Опрашиваем DHT22. Возвращает сразу два флоата (humidity, temperature)
	dhtSensor, err := dht.NewDHT(pinName, dht.Celsius, "")
	if err != nil {
		return SensorData{}, fmt.Errorf("ошибка создания объекта датчика: %w", err)
	}

	humidity, temperature, err := dhtSensor.ReadRetry(3) // 3 попытки, т.к. датчик часто сбоит по вине таймингов Raspberry
	if err != nil {
		return SensorData{}, fmt.Errorf("ошибка чтения DHT22 '%s': %w", d.name, err)
	}

	return SensorData{
		Temperature: temperature,
		Humidity:    humidity,
		Timestamp:   time.Now(),
	}, nil
}
