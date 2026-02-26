package api

import (
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"terrarium-core/internal/automation"
	"terrarium-core/internal/gpio"
	"terrarium-core/internal/storage"

	_ "terrarium-core/docs"
)

// SetupRouter инициализирует движок Gin и принимает все аппаратные и системные зависимости.
func SetupRouter(repo *storage.Repository, relays map[string]gpio.RelayController, engine *automation.Engine) *gin.Engine {
	r := gin.Default()

	// CORS-middleware: разрешаем запросы с фронтенда (Angular dev server и другие origins из .env)
	allowedOrigins := []string{"http://localhost:4200", "http://localhost"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		for _, origin := range strings.Split(envOrigins, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	apiCtrl := &API{
		Repo:   repo,
		Relays: relays,
		Engine: engine,
	}

	// Swagger endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Группа API v1
	v1 := r.Group("/api/v1")
	{
		// Конфигурация и система
		v1.GET("/config", apiCtrl.GetConfig)
		v1.PUT("/config", apiCtrl.UpdateConfig)
		v1.GET("/system/status", apiCtrl.GetSystemStatus)
		v1.POST("/system/mode", apiCtrl.SetSystemMode)

		// Реле (ручное управление)
		v1.GET("/relays", apiCtrl.GetRelays)
		v1.POST("/relays/:id/toggle", apiCtrl.ToggleRelay)

		// Датчики — текущие показания
		v1.GET("/sensors/current", apiCtrl.GetSensorCurrent)

		// Метрики — история датчиков и энергопотребление
		v1.GET("/metrics/sensors", apiCtrl.GetSensorMetrics)
		v1.GET("/metrics/energy", apiCtrl.GetEnergyMetrics)

		// Расписания реле (CRUD)
		v1.GET("/schedules", apiCtrl.GetSchedules)
		v1.POST("/schedules", apiCtrl.CreateSchedule)
		v1.PUT("/schedules/:id", apiCtrl.UpdateSchedule)
		v1.DELETE("/schedules/:id", apiCtrl.DeleteSchedule)

		// Журнал переключений реле
		v1.GET("/relay-logs", apiCtrl.GetRelayLogs)
	}

	return r
}
