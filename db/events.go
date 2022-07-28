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
	Name string `json:"name" db:"name"`
	Date string `json:"date" db:"date"`
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
