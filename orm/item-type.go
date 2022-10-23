package orm

import (
	"reflect"

	"github.com/go-msvc/errors"
	"github.com/google/uuid"
	"github.com/jansemmelink/events/db"
)

type ItemType interface {
	Name() string
	Table() Table
	Create() error
	Add(item interface{}) (string, error)
}

type ormItemType struct {
	name  string
	table *ormTable
}

func (it ormItemType) Name() string {
	return it.name
}

func (it ormItemType) Table() Table {
	return it.table
}

func (it ormItemType) Create() error {
	//exec SQL to create the table
	_, err := db.Db().Exec(it.table.sqlSchema)
	if err != nil {
		return errors.Wrapf(err, "failed to create table %s", it.name)
	}
	return nil
}

func (it ormItemType) Add(item interface{}) (string, error) {
	if reflect.TypeOf(item) != it.table.userStructType {
		return "", errors.Errorf("cannot add %T into table(%s), expecting %v", item, it.name, it.table.userStructType)
	}

	sql := "INSERT INTO `" + it.table.dbName + "` SET"
	args := []interface{}{}

	sep := " "
	id := uuid.New().String()
	sql += sep + "`id`=?"
	args = append(args, id)
	sep = ","

	for i := 0; i < it.Table().DbStructType().NumField(); i++ {
		f := it.Table().DbStructType().Field(i)
		sql += sep + "`" + f.Name + "`=?"
		args = append(args, reflect.ValueOf(item).Field(i))
		sep = ","
	}

	result, err := db.Db().Exec(sql, args...)
	if err != nil {
		return "", errors.Wrapf(err, "failed to add %T", item)
	}
	nr, _ := result.RowsAffected()
	if nr != 1 {
		return "", errors.Errorf("Add(%T) affected %d rows", item, nr)
	}
	return id, nil
}
