package gpio

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// RealDHT22 обеспечивает чтение температуры и влажности с реального диода по 1-wire
type RealDHT22 struct {
	name   string
	pinIdx int
}

// NewRealDHT22 инициализирует аппаратный датчик.
func NewRealDHT22(name string, pinOffset int) (*RealDHT22, error) {
	// Мы используем Python скрипт (adafruit_dht) через exec,
	// так как он работает намного стабильнее на Raspberry Pi 5.

	return &RealDHT22{
		name:   name,
		pinIdx: pinOffset,
	}, nil
}

func (d *RealDHT22) Name() string {
	return d.name
}

func (d *RealDHT22) Read() (SensorData, error) {
	// Вызываем python скрипт из виртуального окружения
	cmd := exec.Command("/opt/venv/bin/python3", "/app/internal/gpio/dht_reader.py", fmt.Sprintf("%d", d.pinIdx))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return SensorData{}, fmt.Errorf("dht_reader.py error: %w, output: %s", err, string(output))
	}

	var result struct {
		Temperature float64 `json:"temperature"`
		Humidity    float64 `json:"humidity"`
		Error       string  `json:"error"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return SensorData{}, fmt.Errorf("json unmarshal failed: %w, raw: %s", err, string(output))
	}

	if result.Error != "" {
		return SensorData{}, fmt.Errorf("dht error: %s", result.Error)
	}

	return SensorData{
		Temperature: result.Temperature,
		Humidity:    result.Humidity,
		Timestamp:   time.Now(),
	}, nil
}
