package db

import (
	"github.com/go-msvc/errors"
)

type EventSummary struct {
	ID   string `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Date string `json:"date" db:"date"`
}

func ListEvents(filter string) ([]EventSummary, error) {
	var rows []EventSummary
	if err := NamedSelect(
		&rows,
		"SELECT id,name,date FROM events WHERE name like :filter",
		map[string]interface{}{
			"filter": "%" + filter + "%",
		}); err != nil {
		return nil, errors.Wrapf(err, "failed to get events")
	}
	return rows, nil
}

type Event struct {
	Name          string `json:"name" db:"name"`
	Description   string `json:"description"`
	StartTime     string `json:"start_time" db:"start_time"`
	EndTime       string `json:"end_time" db:"end_time"`
	LocationID    string `json:"location_id" db:"location_id"`
	Cost          Amount `json:"cost" db:"cost"`
	ParentEventID string `json:"parent_event_id" db:"parent_event_id"`
}

type EventOrganiser struct {
	EventID  string `json:"event_id"`
	PersonID string `json:"person_id"`
	Role     string `json:"role"`
}

type Filter struct {
	Must map[string]interface{}
	Not  map[string]interface{}
}

func GetEvent(id string) (*Event, error) {
	var event Event
	if err := NamedGet(
		&event,
		"SELECT name,date FROM events WHERE id=:id",
		map[string]interface{}{
			"id": id,
		}); err != nil {
		return nil, errors.Wrapf(err, "failed to get event details")
	}
	return &event, nil
}
