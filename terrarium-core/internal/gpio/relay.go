package gpio

import (
	"log"
)

// RelayController описывает интерфейс переключения аппаратного реле.
type RelayController interface {
	// On подает HIGH на пин (включает устройство)
	On() error
	// Off подает LOW на пин (выключает устройство)
	Off() error
	// IsOn возвращает текущее состояние реле
	IsOn() bool
	// Name возвращает имя реле
	Name() string
}

// ==========================================
// MOCK РЕАЛИЗАЦИЯ (для ПК / Windows / Mac)
// ==========================================

// MockRelay имитирует релейный модуль. Реле при старте всегда выключено (безопасность).
type MockRelay struct {
	relayName string
	state     bool
}

// NewMockRelay создает новое "фейковое" реле.
func NewMockRelay(name string) *MockRelay {
	// fail-safe: всегда инициализируем в false (LOW)
	return &MockRelay{
		relayName: name,
		state:     false,
	}
}

func (m *MockRelay) Name() string { return m.relayName }

func (m *MockRelay) On() error {
	if !m.state {
		m.state = true
		log.Printf("[GPIO MOCK] Реле '%s' -> ВКЛЮЧЕНО (HIGH)\n", m.relayName)
	}
	return nil
}

func (m *MockRelay) Off() error {
	if m.state {
		m.state = false
		log.Printf("[GPIO MOCK] Реле '%s' -> ВЫКЛЮЧЕНО (LOW)\n", m.relayName)
	}
	return nil
}

func (m *MockRelay) IsOn() bool {
	return m.state
}
