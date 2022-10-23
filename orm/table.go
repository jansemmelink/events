package orm

import "reflect"

//Table in db
type Table interface {
	Name() string
	UserStructType() reflect.Type
	DbStructType() reflect.Type
}

type ormTable struct {
	userStructType reflect.Type //struct defined in the code at compile time
	dbName         string       //name of database table
	dbStructType   reflect.Type //struct generated to include db tags for read/insert
	sqlSchema      string       //CREATE TABLE `...` (...) ...;
...need mapping from user struct -> db col to iterate over when writing the SQL for INSERT...
}

func (ot ormTable) Name() string {
	return ot.dbName
}

func (ot ormTable) DbStructType() reflect.Type {
	return ot.dbStructType
}

func (ot ormTable) UserStructType() reflect.Type {
	return ot.userStructType
}
