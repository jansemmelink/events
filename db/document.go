package db

type Document struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Revision         int     `json:"revision"`
	Timestamp        SqlTime `json:"timestamp"`
	LoadedByPersonID string  `json:"loaded_by_person_id"`
	ContentID        string  `json:"content_id"`   //in object db
	ContentType      string  `json:"content_type"` //e.g. "pdf", "text/html"
}
