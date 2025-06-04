package apis

import (
	"csv-importer-backend/cmd/csv-importer/model"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type HealthCheckAPI struct {
	db *gorm.DB
}

func NewHealthCheckAPI(db *gorm.DB) *HealthCheckAPI {
	return &HealthCheckAPI{
		db: db,
	}
}

func (a *HealthCheckAPI) Setup(g *echo.Group) {
	g.GET("/healthz", a.healthCheck)
}

func (a *HealthCheckAPI) healthCheck(c echo.Context) error {

	db, err := a.db.DB()
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			model.BaseResponse{
				Message: err.Error(),
			},
		)
	}

	err = db.Ping()
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			model.BaseResponse{
				Message: err.Error(),
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		model.BaseResponse{
			Message: "healthy",
		},
	)
}
