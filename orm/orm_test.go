package orm_test

import (
	"testing"

	"github.com/jansemmelink/events/orm"
)

func TestOneFlatTable(t *testing.T) {
	//basic ORM with only one struct type which is a person item
	//with a name string and age integer...
	type Person struct {
		Name string
		Age  int
	}

	o := orm.New()
	if _, err := o.AddItemType("person", Person{}); err != nil {
		t.Fatalf("failed: %+v", err)
	}

	//validate result:
	// if len(o.ItemTypes()) != 1 {
	// 	t.Logf("Created %d item types != 1", len(o.ItemTypes()))
	// }
	// if len(o.Tables()) != 1 {
	// 	t.Logf("Created %d tables != 1", len(o.Tables()))
	// }

	// //todo: check correct table name and fields
	// tbl, ok := o.Table("person")
	// if !ok {
	// 	t.Fatalf("table person not found")
	// }
	// expectedFieldName := []string{"ID", "Name", "Age"}
	// st := tbl.Type()
	// if len(expectedFieldName) != st.NumField() {
	// 	t.Fatalf("table(%s) has %d != %d fields", tbl.Name(), st.NumField(), len(expectedFieldName))
	// }
	// for i := 0; i < st.NumField(); i++ {
	// 	f := st.Field(i)
	// 	dbName := f.Tag.Get("db")
	// 	if dbName != expectedFieldName[i] {
	// 		t.Fatalf("field[%d].name=%s != %s", i, dbName, expectedFieldName[i])
	// 	}
	// }
}

func Test3(t *testing.T) {
	//orm struct with sub struct in one table
	type B struct {
		Bs string
		Bi int
	}
	type A struct {
		As string
		Ai int
		B  B
	}

	o := orm.New()
	if _, err := o.AddItemType("a", A{}); err != nil {
		t.Fatalf("failed: %+v", err)
	}

}

// func Test2(t *testing.T) {
// 	o := orm.New()
// 	if err := o.Add(Event{}); err != nil {
// 		t.Fatalf("failed: %+v", err)
// 	}

// 	t.Logf("Added %d item types", len(o.ItemTypes()))
// 	for n, it := range o.ItemTypes() {
// 		t.Logf("  %s: %v", n, it.Type())
// 	}

// 	t.Logf("Added %d tables:", len(o.Tables()))
// 	for n, ot := range o.Tables() {
// 		t.Logf("  %s: %v", n, ot.Type())
// 	}
// }

// //Person is item that exist on its own because it has an ID field
// //it is stored in table 'persons'
// type Person struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// 	Age  int    `json:"age"`
// }

// //Event is item that exist on its own because it has an ID field
// //it can be built in an indefinite depth tree because it refers to itself as a parent
// //a person created it
// //a number of persons can be associated with it as organisers
// type Event struct {
// 	ID          string           `json:"id"`
// 	Name        string           `json:"name"`
// 	Description string           `json:"description"`
// 	Creator     *Person          `orm:"required:true" json:"creator"`
// 	Organisers  []EventOrganiser `orm:"min:0;max:0" json:"organisers"`
// 	SubEvents   []Event          `orm:"min:0;max:0" json:"sub_events" doc:"Events or groups inside this event, creates an infinite depth tree"`
// }

// //Event Organiser exists only as part of an event
// //it associates a person with a named role
// type EventOrganiser struct {
// 	Person *Person `json:"person"`
// 	Role   string  `json:"role"`
// }

// //Role can exist on its own because it has an ID
// type Role struct {
// 	ID   string `json:"id"`
// 	Name string `json:"name"`
// }
