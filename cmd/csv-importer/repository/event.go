package repository

import (
	"context"
	"csv-importer-backend/cmd/csv-importer/model"

	"gorm.io/gorm"
)

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{
		db: db,
	}
}

func (r *EventRepo) ListEvents(ctx context.Context) ([]model.Event, error) {

	var events []model.Event

	result := r.db.
		WithContext(ctx).
		Model(&model.Event{}).
		Debug().
		Find(&events)

	if result.Error != nil {
		return nil, result.Error
	}

	return events, nil
}

func (r *EventRepo) CreateEvent(ctx context.Context, event model.Event) error {

	result := r.db.
		WithContext(ctx).
		Model(&event).
		Debug().
		Create(event)

	if result.Error != nil {
		return result.Error
	}

	return nil

}
