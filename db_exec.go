package core

import (
	"database/sql"
	"fmt"
	"reflect"
)

type DBWhere struct {
	FieldName string
	Operator  string
	Value     interface{}
}

/*
* Save data to database
* @param data interface{} Data to save
* @return Error
 */
func SaveDataToDB[T DataBaseObject](ctx Context, data T) Error {
	query, args, insertError := GetInsertQuery(data)
	if insertError != nil {
		ctx.LogError("Error when get insert data = %#v, err = %s", data, insertError.Error())
		return insertError
	}

	ctx.LogInfo("Insert query = %v, args = %v", query, args)
	if _, err := pgSession.ExecContext(ctx, query, args...); err != nil {
		ctx.LogError("Error insert data = %#v, err = %v", data, err)
		return ERROR_INSERT_TO_DB_FAIL
	}

	return nil
}

/*
* Save data to database without primary key
* primary key will be auto increment in database
* @param data interface{} Data to save
* @return Error
 */
func SaveDataToDBWithoutPrimaryKey[T DataBaseObject](ctx Context, data T) Error {
	query, args, pkAddress, insertError := GetInsertQueryWithoutPrimaryKey(data)
	if insertError != nil {
		ctx.LogError("Error when get insert data = %#v, err = %s", data, insertError.Error())
		return insertError
	}

	ctx.LogInfo("Insert query = %v, args = %v", query, args)
	row := pgSession.QueryRowContext(ctx, query, args...)

	err := row.Scan(pkAddress)
	if err != nil {
		ctx.LogError("Get primary key from query fail: %v", err)
		return ERROR_INSERT_TO_DB_FAIL
	}

	return nil
}

/*
* Delete data in database
* @param data interface{} Data to delete
* @return Error
 */
func DeleteDataInDB[T DataBaseObject](ctx Context, data T) Error {
	query, args, deleteError := GetDeleteQuery(data)
	if deleteError != nil {
		ctx.LogError("Error when get delete data = %#v, err = %s", data, deleteError.Error())
		return deleteError
	}

	ctx.LogInfo("Delete query = %v, args = %v", query, args)
	if _, err := pgSession.ExecContext(ctx, query, args...); err != nil {
		ctx.LogError("Error delete data = %#v, err = %v", data, err)
		return ERROR_DB_ERROR
	}

	return nil
}

/*
* Update data in database
* @param data interface{} Data to update
* @return Error
 */
func UpdateDataInDB[T DataBaseObject](ctx Context, data T) Error {
	query, args, updateError := GetUpdateQuery(data)
	if updateError != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, updateError.Error())
		return updateError
	}

	ctx.LogInfo("Update query = %v, args = %v", query, args)
	if _, err := pgSession.ExecContext(ctx, query, args...); err != nil {
		ctx.LogError("Error update data = %#v, err = %s", data, err.Error())
		return ERROR_DB_ERROR
	}

	return nil
}

/*
* Select data from database by primary key
* @param data interface{} Data to select
* @return Error
 */
func SelectById(ctx Context, data DataBaseObject) Error {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return err
	}

	query += fmt.Sprintf(" WHERE %s = $1", data.GetPrimaryKey())
	pk, found := searchPrimaryKey(data)
	if !found {
		ctx.LogError("Error not found primary key = %#v", data)
		return ERROR_DB_ERROR
	}

	ctx.LogInfo("Select query = %v, args = %v", query, pk.Interface())
	row := pgSession.QueryRowContext(ctx, query, pk.Interface())
	if err := row.Scan(params...); err != nil {
		ctx.LogError("Error select data = %#v, err = %v", data, err.Error())
		if err == sql.ErrNoRows {
			return ERROR_NOT_FOUND_IN_DB
		}
		return ERROR_DB_ERROR
	}

	return nil
}

func ListPagingTable(ctx Context, data DataBaseObject, limit int64, offset int64) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	ctx.LogInfo("Select query = %v", query)
	rows, errQuery := pgSession.QueryContext(ctx, query)
	if errQuery != nil {
		ctx.LogError("Error select table %s, err = %s", data.GetTableName(), errQuery.Error())
		return nil, ERROR_DB_ERROR
	}

	// Get list of struct
	resultType := reflect.SliceOf(reflect.TypeOf(data).Elem())
	result := reflect.MakeSlice(resultType, 0, 5)
	for rows.Next() {
		if err := rows.Scan(params...); err != nil {
			ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
			return nil, ERROR_DB_ERROR
		}

		result = reflect.Append(result, reflect.ValueOf(data).Elem())
	}

	return result.Interface(), nil
}

/*
* Select data from database by field: fieldName and fieldValue is passed in parameter
* @return Error
 */
func SelectByField(ctx Context, data DataBaseObject, fieldName string, fieldValue any) Error {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return err
	}

	query += fmt.Sprintf(" WHERE %s = $1", fieldName)

	ctx.LogInfo("Select query = %v, args = %v", query, fieldValue)
	row := pgSession.QueryRowContext(ctx, query, fieldValue)
	if err := row.Scan(params...); err != nil {
		ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
		if err == sql.ErrNoRows {
			return ERROR_NOT_FOUND_IN_DB
		}
		return ERROR_DB_ERROR
	}

	return nil
}

/*
* SelectListByField
* @params: ctx Context, data DataBaseObject, fieldName string, fieldValue any
* @return []DataBaseObject, Error
* @description: select list of data by field
 */
func SelectListByField(ctx Context, data DataBaseObject, fieldName string, fieldValue any) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	query += fmt.Sprintf(" WHERE %s = $1", fieldName)

	ctx.LogInfo("Select query = %v, args = %v", query, fieldValue)
	rows, errQuery := pgSession.QueryContext(ctx, query, fieldValue)
	if errQuery != nil {
		ctx.LogError("Error select %s = %#v, err = %s", fieldName, fieldValue, errQuery.Error())
		return nil, ERROR_DB_ERROR
	}

	// Get list of struct
	resultType := reflect.SliceOf(reflect.TypeOf(data).Elem())
	result := reflect.MakeSlice(resultType, 0, 5)
	for rows.Next() {
		if err := rows.Scan(params...); err != nil {
			ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
			return nil, ERROR_DB_ERROR
		}

		result = reflect.Append(result, reflect.ValueOf(data).Elem())
	}

	return result.Interface(), nil
}

/*
* SelectPagingListByField
* @params: ctx Context, data DataBaseObject, fieldName string, fieldValue any, limit int64, offset int64
* @return []DataBaseObject, Error
* @description: select list of data by field with limit and offset
* @note: this function is used for paging
 */
func SelectPagingListByField(ctx Context, data DataBaseObject, fieldName string, fieldValue any, limit int64, offset int64) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	query += fmt.Sprintf(" WHERE %s = $1 LIMIT %d OFFSET %d", fieldName, limit, offset)

	ctx.LogInfo("Select query = %v, args = %v", query, fieldValue)
	rows, errQuery := pgSession.QueryContext(ctx, query, fieldValue)
	if errQuery != nil {
		ctx.LogError("Error select %s = %#v, err = %s", fieldName, fieldValue, errQuery.Error())
		return nil, ERROR_DB_ERROR
	}

	// Get list of struct
	resultType := reflect.SliceOf(reflect.TypeOf(data).Elem())
	result := reflect.MakeSlice(resultType, 0, 5)
	for rows.Next() {
		if err := rows.Scan(params...); err != nil {
			ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
			return nil, ERROR_DB_ERROR
		}

		result = reflect.Append(result, reflect.ValueOf(data).Elem())
	}

	return result.Interface(), nil
}

/*
* SelectListByFieldsWithCustomOperator
* @params: ctx Context, data DataBaseObject, whereParams ...DBWhere
* @return []DataBaseObject, Error
* @description: select list of data by where params
 */
func SelectListByFieldsWithCustomOperator(ctx Context, data DataBaseObject, whereParams ...DBWhere) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	var args []interface{}
	if len(whereParams) > 0 {
		query += " WHERE "
		args = []interface{}{}
	}

	count := 1
	for _, param := range whereParams {
		if count == 1 {
			query += fmt.Sprintf("%s %s $%d ", param.FieldName, param.Operator, count)
		} else {
			query += fmt.Sprintf("AND %s %s $%d ", param.FieldName, param.Operator, count)
		}

		args = append(args, param.Value)
		count++
	}

	ctx.LogInfo("Select query = %v, args = %#v", query, args)
	rows, errQuery := pgSession.QueryContext(ctx, query, args...)
	if errQuery != nil {
		ctx.LogError("Error select query: %s | args = %#v | err = %s", query, args, errQuery.Error())
		return nil, ERROR_DB_ERROR
	}

	// Get list of struct
	resultType := reflect.SliceOf(reflect.TypeOf(data).Elem())
	result := reflect.MakeSlice(resultType, 0, 5)
	for rows.Next() {
		if err := rows.Scan(params...); err != nil {
			ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
			return nil, ERROR_DB_ERROR
		}

		result = reflect.Append(result, reflect.ValueOf(data).Elem())
	}

	return result.Interface(), nil
}

/*
* SelectListByFields
* @params: ctx Context, data DataBaseObject, mapArgs map[string]interface{}
* @return []DataBaseObject, Error
* @description: select list of data by args
 */
func SelectListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	// Handle where params
	var args []interface{}
	var keys []string
	if len(mapArgs) > 0 {
		query += " WHERE "
		args = []interface{}{}
		keys = []string{}
	}

	count := 1
	for key, value := range mapArgs {
		if count == 1 {
			query += fmt.Sprintf("%s = $%d ", key, count)
		} else {
			query += fmt.Sprintf("AND %s = $%d ", key, count)
		}

		args = append(args, value)
		keys = append(keys, key)
		count++
	}

	ctx.LogInfo("Select query = %v, args = %#v", query, args)
	rows, errQuery := pgSession.QueryContext(ctx, query, args...)
	if errQuery != nil {
		ctx.LogError("Error select %#v = %#v, err = %s", keys, args, errQuery.Error())
		return nil, ERROR_DB_ERROR
	}

	// Get list of struct
	resultType := reflect.SliceOf(reflect.TypeOf(data).Elem())
	result := reflect.MakeSlice(resultType, 0, 5)
	for rows.Next() {
		if err := rows.Scan(params...); err != nil {
			ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
			return nil, ERROR_DB_ERROR
		}

		result = reflect.Append(result, reflect.ValueOf(data).Elem())

	}

	return result.Interface(), nil
}

/*
* SelectPagingListByFields
* @params: ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64
* @return []DataBaseObject, Error
* @description: select list of data by args with limit and offset
* @note: this function is used for paging
 */
func SelectPagingListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	// Handle where params
	var args []interface{}
	var keys []string
	if len(mapArgs) > 0 {
		query += " WHERE "
		args = []interface{}{}
		keys = []string{}
	}

	count := 1
	for key, value := range mapArgs {
		if count == 1 {
			query += fmt.Sprintf("%s = $%d ", key, count)
		} else {
			query += fmt.Sprintf("AND %s = $%d ", key, count)
		}

		args = append(args, value)
		keys = append(keys, key)
		count++
	}

	query = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, limit, offset)

	ctx.LogInfo("Select query = %v, args = %#v", query, args)
	rows, errQuery := pgSession.QueryContext(ctx, query, args...)
	if errQuery != nil {
		ctx.LogError("Error select %#v = %#v, err = %s", keys, args, errQuery.Error())
		return nil, ERROR_DB_ERROR
	}

	// Get list of struct
	resultType := reflect.SliceOf(reflect.TypeOf(data).Elem())
	result := reflect.MakeSlice(resultType, 0, 5)
	for rows.Next() {
		if err := rows.Scan(params...); err != nil {
			ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
			return nil, ERROR_DB_ERROR
		}

		result = reflect.Append(result, reflect.ValueOf(data).Elem())

	}

	return result.Interface(), nil
}

/*
* searchPrimaryKey: search primary key in model
* @params: data DataBaseObject
* @return: reflect.Value, bool
 */
func searchPrimaryKey(data DataBaseObject) (reflect.Value, bool) {
	t, _ := getTypeOfPointer(data)
	found := false
	var idValue reflect.Value

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("db")
		if tag == data.GetPrimaryKey() {
			found = true
			idValue = reflect.ValueOf(data).Elem().FieldByIndex(field.Index)
			break
		}
	}
	return idValue, found
}
