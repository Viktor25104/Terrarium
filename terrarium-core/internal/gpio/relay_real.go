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
	pin.High() // Защита для инвертированных реле: HIGH отключает цепь (Нормально Разомкнуто)

	log.Printf("[GPIO INIT] Аппаратное Реле '%s' инициализировано на пине BCM %d (HIGH/OFF)\n", name, pinNumber)

	return &RealRelay{
		name:  name,
		pin:   pin,
		state: false,
	}, nil
}

func (r *RealRelay) Name() string { return r.name }

func (r *RealRelay) On() error {
	if !r.state {
		r.pin.Low() // LOW включает инвертированное реле (замыкает цепь)
		r.state = true
		log.Printf("[GPIO EVENT] Реле '%s' -> ВКЛЮЧЕНО (LOW)\n", r.name)
	}
	return nil
}

func (r *RealRelay) Off() error {
	if r.state {
		r.pin.High() // HIGH выключает инвертированное реле (размыкает цепь)
		r.state = false
		log.Printf("[GPIO EVENT] Реле '%s' -> ВЫКЛЮЧЕНО (HIGH)\n", r.name)
	}
	return nil
}

func (r *RealRelay) IsOn() bool {
	return r.state
}
