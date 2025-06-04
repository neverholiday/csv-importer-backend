package model

import "time"

type EventStatus string

var (
	Created EventStatus = "draft"
	Start   EventStatus = "start"
	End     EventStatus = "end"
)

type Event struct {
	ID         string      `gorm:"column:id" json:"id"`
	Name       string      `gorm:"column:name" json:"name"`
	Status     EventStatus `gorm:"column:status" json:"status"`
	CreateDate time.Time   `gorm:"column:create_date" json:"create_date"`
	UpdateDate time.Time   `gorm:"column:update_date" json:"update_date"`
	DeleteDate *time.Time  `gorm:"column:delete_date" json:"delete_date,omitempty"`
}

func (m *Event) TableName() string {
	return "events"
}

type TodoEvent struct {
	ID         string     `gorm:"column:id" json:"id"`
	EventID    string     `gorm:"column:event_id" json:"event_id"`
	CreateDate time.Time  `gorm:"column:create_date" json:"create_date"`
	UpdateDate time.Time  `gorm:"column:update_date" json:"update_date"`
	DeleteDate *time.Time `gorm:"column:delete_date" json:"delete_date,omitempty"`
}
