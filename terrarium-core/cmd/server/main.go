package main

import (
	"context"
	"log"
	"os"

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

	// 4. Инициализация Аппаратуры (GPIO) - ТОЛЬКО БЕЗ МОКОВ
	log.Println("[СТАРТ] Инициализация БОЕВОГО оборудования Raspberry Pi 5 (libgpiod)...")

	var warmSensor, coldSensor gpio.SensorReader
	var relays map[string]gpio.RelayController
	// В будущем брать пины из .env GPIO_MAPPING, пока хардкод для стабильности
	warmSensor, err = gpio.NewRealDHT22("WarmZone", 4) // GPIO 4 (D4 / D5?)
	if err != nil {
		log.Printf("[ВНИМАНИЕ] Ошибка инициализации DHT22 (Warm): %v\n", err)
		// Программа не должна падать, если датчик временно отвалился
	}

	coldSensor, err = gpio.NewRealDHT22("ColdZone", 17) // GPIO 17
	if err != nil {
		log.Printf("[ВНИМАНИЕ] Ошибка инициализации DHT22 (Cold): %v\n", err)
	}

	relays = make(map[string]gpio.RelayController)

	relays["heat_mat"], err = gpio.NewRealRelay("heat_mat", 22) // Оранжевый
	if err != nil {
		log.Fatalf("Ошибка Реле heat_mat: %v", err)
	}

	relays["fogger"], err = gpio.NewRealRelay("fogger", 27) // Серый (из питона BCM 27)
	if err != nil {
		log.Fatalf("Ошибка Реле fogger: %v", err)
	}

	relays["light"], err = gpio.NewRealRelay("light", 17) // Белый (из питона BCM 17)
	if err != nil {
		log.Fatalf("Ошибка Реле light: %v", err)
	}

	relays["spare"], err = gpio.NewRealRelay("spare", 23) // Фиолетовый (BCM 23)
	if err != nil {
		log.Fatalf("Ошибка Реле spare: %v", err)
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
