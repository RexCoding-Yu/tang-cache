package cache

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"reflect"
)

func isBasicType(k reflect.Kind) bool {
	return (k > 0 && k < reflect.Array) || (k == reflect.String)
}

func getObjectsAfterLoad(db *gorm.DB) (primaryKeys []string, objects []interface{}) {
	primaryKeys = make([]string, 0)
	values := make([]reflect.Value, 0)
	isPluck := false

	destValue := reflect.Indirect(reflect.ValueOf(db.Statement.Dest))
	switch destValue.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < destValue.Len(); i++ {
			elem := destValue.Index(i)
			values = append(values, elem)
		}
		if isBasicType(destValue.Type().Elem().Kind()) {
			isPluck = true
		}
	case reflect.Struct:
		values = append(values, destValue)
	}

	var valueOf func(context.Context, reflect.Value) (value interface{}, zero bool) = nil
	if db.Statement.Schema != nil {
		for _, field := range db.Statement.Schema.Fields {
			if field.PrimaryKey {
				valueOf = field.ValueOf
				break
			}
		}
	}

	objects = make([]interface{}, 0, len(values))
	for _, elemValue := range values {
		if valueOf != nil && !isPluck {
			primaryKey, isZero := valueOf(context.Background(), elemValue)
			if isZero {
				continue
			}
			primaryKeys = append(primaryKeys, fmt.Sprintf("%v", primaryKey))
		}
		objects = append(objects, elemValue.Interface())
	}
	return primaryKeys, objects
}
