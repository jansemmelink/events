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

type NewEventRequest struct {
	Name string `json:"name"`
	Date string `json:"data"`
}

func AddEvent(NewEventRequest) (*Event, error) {
	...get auth details to know who is the organiser
	//allow multiple persons to be added/removed as organisers
	//create list of event contacts, e.g. "organisers":..., "enquiries":..., "admin":... and allow them to edit the list
	//need ultimately to grant them individually access to event operations, but can do that later because will need profiles of what is allowed for role of helper.
	return nil, errors.Errorf("NYI")
}
