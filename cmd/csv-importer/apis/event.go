package apis

import (
	"context"
	"csv-importer-backend/cmd/csv-importer/model"
	"net/http"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/goforj/godump"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type IEventRepo interface {
	ListEvents(ctx context.Context) ([]model.Event, error)
	CreateEvent(ctx context.Context, event model.Event) error
}

type EventAPI struct {
	eventRepo IEventRepo
}

func NewEventAPI(eventRepo IEventRepo) *EventAPI {

	return &EventAPI{
		eventRepo: eventRepo,
	}
}

func (a *EventAPI) Setup(g *echo.Group) {
	g.GET("/events", a.listEvents)
	g.POST("/event", a.createEvent)
}

func (a *EventAPI) listEvents(c echo.Context) error {

	ctx := c.Request().Context()

	events, err := a.eventRepo.ListEvents(ctx)
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
			Message: "success",
			Data:    events,
		},
	)
}

func (a *EventAPI) createEvent(c echo.Context) error {

	ctx := c.Request().Context()

	eventName := c.FormValue("name")
	csvfile, err := c.FormFile("csvfile")

	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			model.BaseResponse{
				Message: err.Error(),
			},
		)
	}

	cf, err := csvfile.Open()
	if err != nil {
		return c.JSON(
			http.StatusBadRequest,
			model.BaseResponse{
				Message: err.Error(),
			},
		)
	}

	defer cf.Close()

	var todos []model.TodoCSV
	err = gocsv.Unmarshal(cf, &todos)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			model.BaseResponse{
				Message: err.Error(),
			},
		)

	}

	godump.Dump(todos)

	id, err := uuid.NewV7()
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			model.BaseResponse{
				Message: err.Error(),
			},
		)
	}

	req := model.EventCreateRequest{
		Name: eventName,
	}

	event := model.Event{
		ID:         id.String(),
		Name:       req.Name,
		Status:     model.Created,
		CreateDate: time.Now(),
		UpdateDate: time.Now(),
	}
	err = a.eventRepo.CreateEvent(
		ctx,
		event,
	)

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
			Message: "success",
			Data:    event,
		},
	)
}
