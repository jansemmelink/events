package main

import (
	"fmt"

	"github.com/jansemmelink/events/orm"
	"github.com/stewelarend/logger"
)

var log = logger.New().WithLevel(logger.LevelDebug)

func main() {
	o := orm.New()
	type Person struct {
		Name string
		Age  int
	}
	personItemType, err := o.AddItemType("person", Person{})
	if err != nil {
		panic(err)
	}
	if err := personItemType.Create(); err != nil {
		panic(fmt.Sprintf("failed to create table: %+v", err))
	}

	id, err := personItemType.Add(Person{Name: "Jan", Age: 10})
	if err != nil {
		panic(fmt.Sprintf("failed to add p1: %+v", err))
	}
	log.Debugf("Added %s", id)
}
