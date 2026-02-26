package gpio

import (
	"fmt"
	"log"

	"github.com/stianeikeland/go-rpio/v4"
)

// RealRelay управляет настоящим реле через пины Raspberry Pi (используя /dev/gpiomem)
type RealRelay struct {
	name  string
	pin   rpio.Pin
	state bool // Кэшируем состояние
}

// NewRealRelay инициализирует физический пин как Output.
func NewRealRelay(name string, pinNumber int) (*RealRelay, error) {
	// rpio.Open() должен вызываться один раз на всё приложение (обычно в main.go),
	// но библиотека go-rpio безопасно обрабатывает многократные вызовы (ref counting).
	if err := rpio.Open(); err != nil {
		return nil, fmt.Errorf("ошибка инициализации /dev/gpiomem: %w", err)
	}

	pin := rpio.Pin(pinNumber)
	pin.Output()
	pin.Low() // Безопасный старт - выключаем реле

	log.Printf("[GPIO INIT] Аппаратное Реле '%s' инициализировано на пине BCM %d (LOW)\n", name, pinNumber)

	return &RealRelay{
		name:  name,
		pin:   pin,
		state: false,
	}, nil
}

func (r *RealRelay) Name() string { return r.name }

func (r *RealRelay) On() error {
	if !r.state {
		r.pin.High()
		r.state = true
		log.Printf("[GPIO EVENT] Реле '%s' -> ВКЛЮЧЕНО (HIGH)\n", r.name)
	}
	return nil
}

func (r *RealRelay) Off() error {
	if r.state {
		r.pin.Low()
		r.state = false
		log.Printf("[GPIO EVENT] Реле '%s' -> ВЫКЛЮЧЕНО (LOW)\n", r.name)
	}
	return nil
}

func (r *RealRelay) IsOn() bool {
	return r.state
}
