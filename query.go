package core

import (
	"fmt"
	"reflect"
	"strings"
)

type DataBaseObject interface {
	GetTableName() string
	GetPrimaryKey() string
}

func splitPrimaryKey(model DataBaseObject) ([]string, int) {
	if model.GetPrimaryKey() == BLANK {
		return []string{}, 0
	}

	primaryKeys := model.GetPrimaryKey()
	primaryKeyFields := strings.Split(primaryKeys, ",")
	return primaryKeyFields, len(primaryKeyFields)
}

/*
* Get select query: generate a select query from a model
* @params: model DataBaseObject
* @return: string, Error
 */
func GetSelectQuery[T DataBaseObject](model T) (string, []any, Error) {
	t, err := getTypeOfPointer(model)
	if err != nil {
		return BLANK, nil, err
	}
	v := reflect.ValueOf(model).Elem()

	// Get table name from model
	query := "SELECT "
	tableName := model.GetTableName()

	numField := t.NumField()
	dbFieldLength := 0
	scanParams := []any{}
	for i := 0; i < numField; i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")

		if tag == BLANK {
			continue
		}

		if i == 0 {
			query += tag
		} else {
			query += ", " + tag
		}
		scanParams = append(scanParams, v.Field(i).Addr().Interface())
		dbFieldLength++
	}

	query += " "

	if dbFieldLength == 0 {
		return BLANK, nil, ERROR_MODEL_HAVE_NO_FIELD
	}

	if query[len(query)-2:] == ", " {
		query = query[:len(query)-1]
	}

	query += "FROM " + tableName
	return query, scanParams, nil
}

/*
* Get insert query: generate an insert query from a model
* @params: model DataBaseObject
* @return: string, []interface{}, Error
 */
func GetInsertQuery[T DataBaseObject](model T) (string, []any, Error) {
	t, err := getTypeOfPointer(model)
	if err != nil {
		return BLANK, nil, err
	}
	v := reflect.ValueOf(model).Elem()

	// Generate insert query
	tableName := model.GetTableName()
	fields := BLANK
	questionString := BLANK
	args := []any{}
	count := 1
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")

		if tag == BLANK {
			continue
		}

		if i != t.NumField()-1 {
			fields += tag + ","
			questionString += fmt.Sprintf("$%d,", count)
		} else {
			fields += tag
			questionString += fmt.Sprintf("$%d", count)
		}

		count++
		args = append(args, v.Field(i).Interface())
	}

	if len(fields) == 0 {
		return BLANK, nil, ERROR_MODEL_HAVE_NO_FIELD
	}

	if fields[len(fields)-1:] == "," {
		fields = fields[:len(fields)-1]
		questionString = questionString[:len(questionString)-1]
	}

	query := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", tableName, fields, questionString)

	return query, args, nil
}

/*
* Get insert query: generate an insert query from a model without insert primary key
* @params: model DataBaseObject
* @return: string, []interface{}, Error
 */
func GetInsertQueryWithoutPrimaryKey[T DataBaseObject](model T) (string, []any, any, Error) {
	t, err := getTypeOfPointer(model)
	if err != nil {
		return BLANK, nil, nil, err
	}
	v := reflect.ValueOf(model).Elem()

	// Generate insert query

	tableName := model.GetTableName()
	// Primary key
	primaryKeys, numPrimaryKeys := splitPrimaryKey(model)
	if numPrimaryKeys > 1 {
		return BLANK, nil, nil, ERROR_NO_SUPPORT_FOR_MANY_PRIMARY_KEYS
	} else if numPrimaryKeys == 0 {
		return BLANK, nil, nil, ERROR_NOT_FOUND_PRIMARY_KEY
	}

	var primaryKeyAddress interface{}
	foundPrimaryKey := false

	fields := BLANK
	questionString := BLANK
	args := []any{}
	count := 1
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")

		if tag == BLANK {
			continue
		}

		if tag == primaryKeys[0] {
			primaryKeyAddress = v.Field(i).Addr().Interface()
			foundPrimaryKey = true
			continue
		}

		if i != t.NumField()-1 {
			fields += tag + ","
			questionString += fmt.Sprintf("$%d,", count)
		} else {
			fields += tag
			questionString += fmt.Sprintf("$%d", count)
		}

		count++
		args = append(args, v.Field(i).Interface())
	}

	if len(fields) == 0 {
		return BLANK, nil, primaryKeyAddress, ERROR_MODEL_HAVE_NO_FIELD
	}

	if !foundPrimaryKey {
		return BLANK, nil, primaryKeyAddress, ERROR_NOT_FOUND_PRIMARY_KEY
	}

	if fields[len(fields)-1:] == "," {
		fields = fields[:len(fields)-1]
		questionString = questionString[:len(questionString)-1]
	}

	query := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING %s", tableName, fields, questionString, primaryKeys[0])

	return query, args, primaryKeyAddress, nil
}

/*
* Get update query: generate an update query from a model
* @params: model DataBaseObject
* @return: string, []any, Error
 */
func GetUpdateQuery[T DataBaseObject](model T) (string, []any, Error) {
	t, err := getTypeOfPointer(model)
	if err != nil {
		return BLANK, nil, err
	}
	v := reflect.ValueOf(model).Elem()

	tableName := model.GetTableName()
	primaryKeys, numPrimaryKeys := splitPrimaryKey(model)
	if numPrimaryKeys == 0 {
		return BLANK, nil, ERROR_NOT_FOUND_PRIMARY_KEY
	}
	primaryValues := make([]interface{}, len(primaryKeys))

	var setString string
	var args []interface{}
	count := 1
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")

		if tag == BLANK {
			continue
		}

		isPrimaryKey := false
		for k, key := range primaryKeys {
			if tag == key {
				isPrimaryKey = true
				primaryValues[k] = v.Field(i).Interface()
			}
		}

		if isPrimaryKey {
			continue
		}

		value := v.Field(i).Interface()

		if i != t.NumField()-1 {
			setString += fmt.Sprintf("%s = $%d, ", tag, count)
		} else {
			setString += fmt.Sprintf("%s = $%d", tag, count)
		}
		count++
		args = append(args, value)
	}

	// Check argument and primary key
	if len(args) == 0 {
		return BLANK, nil, ERROR_MODEL_HAVE_NO_FIELD
	}
	if len(primaryValues) == 0 || len(primaryKeys) != len(primaryValues) {
		return BLANK, nil, ERROR_NOT_FOUND_PRIMARY_KEY
	}

	if setString[len(setString)-2:] == ", " {
		setString = setString[:len(setString)-2]
	}

	query := fmt.Sprintf("UPDATE %s SET %s", tableName, setString)
	for i, key := range primaryKeys {
		if i == 0 {
			query += fmt.Sprintf(" WHERE %s = $%d", key, count)
		} else {
			query += fmt.Sprintf(" AND %s = $%d", key, count+i)
		}
	}

	args = append(args, primaryValues...)

	return query, args, nil
}

/*
* Get delete query: generate a delete query from a model
* @params: model DataBaseObject
* @return: string, []any, error
 */
func GetDeleteQuery[T DataBaseObject](model T) (string, []any, Error) {
	// Check model is pointer of struct
	_, err := getTypeOfPointer(model)
	if err != nil {
		return BLANK, nil, err
	}

	tableName := model.GetTableName()
	pkValues, found := searchPrimaryKey(model)
	if !found {
		return BLANK, nil, ERROR_NOT_FOUND_PRIMARY_KEY
	}

	primaryKeys, number := splitPrimaryKey(model)
	if number != len(pkValues) {
		return BLANK, nil, ERROR_NOT_FOUND_PRIMARY_KEY
	}

	query := fmt.Sprintf("DELETE FROM %s", tableName)
	for i, key := range primaryKeys {
		if i == 0 {
			query += fmt.Sprintf(" WHERE %s = $%d", key, i+1)
		} else {
			query += fmt.Sprintf(" AND %s = $%d", key, i+1)
		}
	}

	return query, pkValues, nil
}

/*
* searchPrimaryKey: search primary key in model
* @params: data DataBaseObject
* @return: reflect.Value, bool
 */
func searchPrimaryKey(data DataBaseObject) ([]any, bool) {
	t, _ := getTypeOfPointer(data)
	found := false
	var idValues []any
	primaryKeys, numPrimaryKeys := splitPrimaryKey(data)
	if numPrimaryKeys == 0 {
		return nil, false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		for _, key := range primaryKeys {
			if tag == key {
				idValues = append(idValues, reflect.ValueOf(data).Elem().FieldByIndex(field.Index).Interface())
				break
			}
		}
	}

	if len(idValues) == numPrimaryKeys {
		found = true
	}
	return idValues, found
}
