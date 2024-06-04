package core

import (
	"fmt"
	"reflect"
)

type DataBaseObject interface {
	GetTableName() string
	GetPrimaryKey() string
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
	primaryKey := model.GetPrimaryKey()
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

		if tag == primaryKey {
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

	query := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING %s", tableName, fields, questionString, primaryKey)

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
	primaryKey := model.GetPrimaryKey()
	var primaryValue interface{}

	var setString string
	var args []interface{}
	count := 1
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")

		if tag == BLANK {
			continue
		}

		if tag == primaryKey {
			primaryValue = v.Field(i).Interface()
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

	if len(args) == 0 {
		return BLANK, nil, ERROR_MODEL_HAVE_NO_FIELD
	}
	if primaryValue == nil {
		return BLANK, nil, ERROR_NOT_FOUND_PRIMARY_KEY
	}

	if setString[len(setString)-1:] == "," {
		setString = setString[:len(setString)-1]
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = $%d", tableName, setString, primaryKey, count)
	args = append(args, primaryValue)

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
	pkValue, found := searchPrimaryKey(model)
	if !found {
		return BLANK, nil, ERROR_NOT_FOUND_PRIMARY_KEY
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", tableName, model.GetPrimaryKey())
	args := []any{pkValue.Interface()}

	return query, args, nil
}
