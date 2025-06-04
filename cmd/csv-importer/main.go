package main

import (
	"csv-importer-backend/cmd/csv-importer/apis"
	"csv-importer-backend/cmd/csv-importer/repository"
	"fmt"
	"os"

	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type EnvCfg struct {
	DBHost     string `envconfig:"DB_HOST" required:"true"`
	DBPort     int    `envconfig:"DB_PORT" required:"true"`
	DBUser     string `envconfig:"DB_USER" required:"true"`
	DBPassword string `envconfig:"DB_PASSWORD" required:"true"`
	DBName     string `envconfig:"DB_NAME" required:"true"`
}

func main() {

	err := os.Setenv("TZ", "UTC")
	if err != nil {
		panic(err)
	}

	var cfg EnvCfg
	err = envconfig.Process("CSV_IMPORTER", &cfg)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(
		postgres.Open(
			fmt.Sprintf(
				"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
				cfg.DBHost,
				cfg.DBPort,
				cfg.DBUser,
				cfg.DBPassword,
				cfg.DBName,
			),
		),
	)

	if err != nil {
		panic(err)
	}

	e := echo.New()

	rootg := e.Group("")
	v1g := rootg.Group("/api/v1")

	apis.
		NewHealthCheckAPI(db).
		Setup(rootg)

	eventRepo := repository.NewEventRepo(db)

	apis.
		NewEventAPI(eventRepo).
		Setup(v1g)

	e.Start(":8080")

}
