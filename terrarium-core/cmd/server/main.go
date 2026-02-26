package main

import (
	"context"
	"log"
	"os"
	"strings"

	"terrarium-core/internal/api"
	"terrarium-core/internal/automation"
	"terrarium-core/internal/gpio"
	"terrarium-core/internal/storage"

	"github.com/joho/godotenv"
)

// @title API Платформы Климат-Контроля Террариума (Terrarium Climate)
// @version 1.0.0
// @description Экстенсивная (excessive) документация API системы управления микроклиматом террариума на базе Raspberry Pi 5.
// @host localhost:8080
// @BasePath /

func main() {
	log.Println("Запуск ядра автоматизации климата (Terrarium Core)...")

	// 1. Загрузка конфигурации окружения
	_ = godotenv.Load("../.env")

	// 2. Инициализация глобального контекста
	ctx := context.Background()

	// 3. Подключение к БД
	db, err := storage.Connect(ctx)
	if err != nil {
		log.Fatalf("Критическая ошибка инициализации БД: %v", err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)

	// 4. Инициализация Аппаратуры (GPIO) с переключением Mock/Real
	isMock := true // По умолчанию безопасно используем моки
	if mockEnv := strings.ToLower(os.Getenv("MOCK_HARDWARE")); mockEnv == "false" || mockEnv == "0" {
		isMock = false
	}

	var warmSensor, coldSensor gpio.SensorReader
	var relays map[string]gpio.RelayController

	if isMock {
		log.Println("[СТАРТ] Инициализация программных ЗАГЛУШЕК (Mock) оборудования (ПК режим)...")
		warmSensor = gpio.NewMockDHT22("WarmZone", 32.5, 60.0)
		coldSensor = gpio.NewMockDHT22("ColdZone", 25.0, 70.0)

		relays = map[string]gpio.RelayController{
			"heat_mat": gpio.NewMockRelay("heat_mat"),
			"fogger":   gpio.NewMockRelay("fogger"),
			"light":    gpio.NewMockRelay("light"),
			"spare":    gpio.NewMockRelay("spare"),
		}
	} else {
		log.Println("[СТАРТ] Инициализация БОЕВОГО оборудования Raspberry Pi 5 (libgpiod)...")
		// Для реального запуска парсим номера пинов из GPIO_MAPPING (json) в .env
		// В данном примере хардкодим пины для простоты, но в бою берем из Config/Env

		var err error
		warmSensor, err = gpio.NewRealDHT22("WarmZone", 4) // GPIO 4
		if err != nil {
			log.Fatalf("Ошибка DHT22 (Warm): %v", err)
		}

		coldSensor, err = gpio.NewRealDHT22("ColdZone", 17) // GPIO 17
		if err != nil {
			log.Fatalf("Ошибка DHT22 (Cold): %v", err)
		}

		relays = make(map[string]gpio.RelayController)

		relays["heat_mat"], err = gpio.NewRealRelay("heat_mat", 22)
		if err != nil {
			log.Fatalf("Ошибка Реле: %v", err)
		}

		relays["fogger"], err = gpio.NewRealRelay("fogger", 23)
		if err != nil {
			log.Fatalf("Ошибка Реле: %v", err)
		}

		relays["light"], err = gpio.NewRealRelay("light", 24)
		if err != nil {
			log.Fatalf("Ошибка Реле: %v", err)
		}

		relays["spare"], err = gpio.NewRealRelay("spare", 25)
		if err != nil {
			log.Fatalf("Ошибка Реле: %v", err)
		}
	}

	// 5. Запуск фонового движка автоматизации (Конечного Автомата)
	engine := automation.NewEngine(
		repo,
		warmSensor,
		coldSensor,
		relays["heat_mat"],
		relays["fogger"],
		relays["light"],
	)

	// Горутина автоматизации начинает работу в фоне
	go engine.Start(ctx)

	// 6. Настройка HTTP Роутинга и Swagger
	router := api.SetupRouter(repo, relays, engine)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("HTTP Сервер запущен. Swagger: http://localhost:%s/swagger/index.html\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска HTTP сервера: %v", err)
	}
}
