package orm

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/go-msvc/errors"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func New() Orm {
	return &orm{
		itemType: map[string]*ormItemType{},
		table:    map[string]*ormTable{},
	}
}

type Orm interface {
	AddItemType(name string, tmpl interface{}) (ItemType, error)
	// ItemTypes() map[string]ItemType
	// Tables() map[string]Table
	// Table(name string) (Table, bool)

	Add(item interface{}) (string, error)
}

type orm struct {
	itemType map[string]*ormItemType
	table    map[string]*ormTable
}

//AddItemType() adds an item type to the ORM
func (o *orm) AddItemType(name string, tmpl interface{}) (ItemType, error) {
	//todo: validate name

	if _, ok := o.itemType[name]; ok {
		return nil, errors.Errorf("duplicate item type(%s)", name)
	}

	t := reflect.TypeOf(tmpl)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, errors.Errorf("%v is not a struct", t)
	}

	//prepare as struct type in Go for storing this item in a database table
	//any complex structure will result in sub-tables, so multiple tables
	//may be created here
	tables, err := o.structTables("item_"+name, t)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create db structs for %T", tmpl)
	}
	log.Debugf("tables: %+v", tables)
	it := &ormItemType{
		name:  name,
		table: tables[0],
	}
	o.itemType[name] = it
	return it, nil
} //orm.Add()

//make list of tables needed to store the struct
func (o *orm) structTables(dbName string, t reflect.Type) ([]*ormTable, error) {
	tables := []*ormTable{}

	//start a new table to store this struct
	structTable := &ormTable{
		dbName:         dbName,
		userStructType: t,
	}

	dbStructFields := []reflect.StructField{}
	colDefs := []string{}
	colNames := map[string]bool{}

	//for item table: start with ID
	//if ... {
	dbStructFields = append(dbStructFields, reflect.StructField{
		Name: "ID",
		Type: reflect.TypeOf(""),
		Tag:  "db:\"id\"",
	})
	colDefs = append(colDefs, "`id` varchar(40) DEFAULT uuid() NOT NULL")
	colNames["id"] = true
	//}//if item

	//add all data fields
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		ft := f.Type

		//struct field name must be Go public, i.e. CamelCase
		fieldName := f.Name

		//db column name is snake_case
		dbColName := snakeCase(fieldName)
		if _, ok := colNames[dbColName]; ok {
			return nil, errors.Errorf("%s.%s makes duplicate colName(%s)", t.String(), f.Name, dbColName)
		}

		switch ft.Kind() {
		case reflect.String:
			colDefs = append(colDefs, fmt.Sprintf("`%s` varchar(128) NOT NULL", dbColName))
		case reflect.Int:
			colDefs = append(colDefs, fmt.Sprintf("`%s` int(11) NOT NULL", dbColName))
		case reflect.Struct:
			//sub struct is added to the base table
			colDefs = append(colDefs, fmt.Sprintf("`%s` todo() NOT NULL", dbColName))
		// case reflect.Ptr:
		//ptr to other item is a item reference, storing only the item id
		//ptr to non-item struct is an optional sub-struct stored in the same table with null values
		// 	colDefs = append(colDefs, fmt.Sprintf("`%s` todo() NOT NULL", dbColName))
		// case reflect.Slice:
		// 	colDefs = append(colDefs, fmt.Sprintf("`%s` todo() NOT NULL", dbColName))
		default:
			return nil, errors.Errorf("%s.%s has unexpected kind %v", t.String(), f.Name, ft.Kind())
		}
	} //for each field

	if len(dbStructFields) == 0 {
		return nil, errors.Errorf("struct type %s has no db fields", t.String())
	}
	structTable.dbStructType = reflect.StructOf(dbStructFields)

	structTable.sqlSchema = "CREATE TABLE `" + dbName + "` ("
	for i, colDef := range colDefs {
		if i > 0 {
			structTable.sqlSchema += ","
		}
		structTable.sqlSchema += colDef
	}
	structTable.sqlSchema += ")"

	tables = append(tables, structTable)
	return tables, nil
} //orm.structTables()

func (o orm) ItemTypes() map[string]ItemType {
	m := map[string]ItemType{}
	for n, it := range o.itemType {
		m[n] = it
	}
	return m
}

func (o orm) Tables() map[string]Table {
	m := map[string]Table{}
	// for n, tbl := range o.tableByName {
	// 	m[n] = tbl
	// }
	return m
}

func (o orm) Table(name string) (Table, bool) {
	// if tbl, ok := o.tableByName[name]; ok {
	// 	return tbl, true
	// }
	return nil, false
}

func snakeCase(camelStr string) string {
	l := len(camelStr)
	s := ""
	for i, c := range camelStr {
		s += strings.ToLower(string(c))
		if i < l-1 && unicode.IsLower(c) && unicode.IsUpper(c) {
			s += "_"
		}
	}
	return s
}

func (o orm) Add(item interface{}) (string, error) {
	t := reflect.TypeOf(item)
	it, ok := o.itemType[t.String()]
	if !ok {
		return "", errors.Errorf("%T is not an item type", item)
	}
	return it.Add(item)
}
