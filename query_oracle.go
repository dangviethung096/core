package core

import (
	"database/sql"
	"fmt"
	"reflect"
)

/*
* Get insert query for oracle: generate an insert query from a model
* @params: model DataBaseObject
* @return: string, map[string]any, Error
 */
func GetInsertQueryForOracle[T DataBaseObject](model T) (string, []any, Error) {
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
			questionString += fmt.Sprintf(":%s,", tag)
		} else {
			fields += tag
			questionString += fmt.Sprintf(":%s", tag)
		}

		count++
		args = append(args, sql.Named(tag, v.Field(i).Interface()))
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
* Get insert query for oracle: generate an insert query from a model without insert primary key
* @params: model DataBaseObject
* @return: string, map[string]any, Error
 */
func GetInsertQueryWithoutPrimaryKeyForOracle[T DataBaseObject](model T) (string, []any, any, Error) {
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
			questionString += fmt.Sprintf(":%s,", tag)
		} else {
			fields += tag
			questionString += fmt.Sprintf(":%s", tag)
		}

		count++
		args = append(args, sql.Named(tag, v.Field(i).Interface()))
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

	query := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s) RETURNING %s INTO :%s", tableName, fields, questionString, primaryKeys[0], primaryKeys[0])

	// Append primary key to args
	args = append(args, sql.Named(primaryKeys[0], primaryKeyAddress))

	return query, args, primaryKeyAddress, nil
}

/*
* Get delete query for oracle: generate a delete query from a model
* @params: model DataBaseObject
* @return: string, []any, error
 */
func GetDeleteQueryForOracle[T DataBaseObject](model T) (string, []any, Error) {
	// Check model is pointer of struct
	_, err := getTypeOfPointer(model)
	if err != nil {
		return BLANK, nil, err
	}

	tableName := model.GetTableName()
	pkValues, found := listPrimaryKey(model)
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
			query += fmt.Sprintf(" WHERE %s = :%s", key, key)
		} else {
			query += fmt.Sprintf(" AND %s = :%s", key, key)
		}
	}

	return query, pkValues, nil
}

/*
* Get update query: generate an update query from a model
* @params: model DataBaseObject
* @return: string, map[string]any, Error
 */
func GetUpdateQueryForOracle[T DataBaseObject](model T) (string, []any, Error) {
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
	primaryValues := map[string]any{}

	var setString string
	args := []any{}
	count := 1
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")

		if tag == BLANK {
			continue
		}

		isPrimaryKey := false
		for _, key := range primaryKeys {
			if tag == key {
				isPrimaryKey = true
				primaryValues[tag] = v.Field(i).Interface()
			}
		}

		if isPrimaryKey {
			continue
		}

		value := v.Field(i).Interface()

		if i != t.NumField()-1 {
			setString += fmt.Sprintf("%s = :%s, ", tag, tag)
		} else {
			setString += fmt.Sprintf("%s = :%s", tag, tag)
		}
		count++
		args = append(args, sql.Named(tag, value))
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
			query += fmt.Sprintf(" WHERE %s = :%s", key, key)
		} else {
			query += fmt.Sprintf(" AND %s = :%s", key, key)
		}

		args = append(args, sql.Named(key, primaryValues[key]))
	}

	return query, args, nil
}

/*
* listPrimaryKey: search primary key in model
* @params: data DataBaseObject
* @return: map[string]any, bool
 */
func listPrimaryKey(data DataBaseObject) ([]any, bool) {
	t, _ := getTypeOfPointer(data)
	found := false
	idValues := []any{}
	primaryKeys, numPrimaryKeys := splitPrimaryKey(data)
	if numPrimaryKeys == 0 {
		return nil, false
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		for _, key := range primaryKeys {
			if tag == key {
				idValues = append(idValues, sql.Named(key, reflect.ValueOf(data).Elem().FieldByIndex(field.Index).Interface()))
				break
			}
		}
	}

	if len(idValues) == numPrimaryKeys {
		found = true
	}
	return idValues, found
}
