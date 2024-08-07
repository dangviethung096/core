package core

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/lib/pq"
)

type DBWhere struct {
	FieldName string
	Operator  string
	Value     interface{}
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
		pqError, ok := err.(*pq.Error)
		if ok {
			if pqError.Code.Name() == DB_ERROR_NAME_UNIQUE_VIOLATION {
				return ERROR_DB_UNIQUE_VIOLATION
			} else if pqError.Code.Name() == DB_ERROR_NAME_FOREIGN_KEY_VIOLATION {
				return ERROR_DB_FOREIGN_KEY_VIOLATION
			}
		}
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
		ctx.LogError("Error insert data = %#v, err = %v", data, err)
		pqError, ok := err.(*pq.Error)
		if ok {
			if pqError.Code.Name() == "unique_violation" {
				return ERROR_DB_UNIQUE_VIOLATION
			} else if pqError.Code.Name() == "foreign_key_violation" {
				return ERROR_DB_FOREIGN_KEY_VIOLATION
			}
		}
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
		pqError, ok := err.(*pq.Error)
		if ok {
			if pqError.Code.Name() == DB_ERROR_NAME_FOREIGN_KEY_VIOLATION {
				return ERROR_DB_FOREIGN_KEY_VIOLATION
			}
		}
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return nil
}

func DeleteDataWithWhereQuery[T DataBaseObject](ctx Context, data T, whereQuery string) Error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", data.GetTableName(), whereQuery)

	ctx.LogInfo("Delete query = %v", query)
	ret, err := pgSession.ExecContext(ctx, query)
	if err != nil {
		ctx.LogError("Error delete data = %#v, err = %v", data, err)
		pqError, ok := err.(*pq.Error)
		if ok {
			if pqError.Code.Name() == DB_ERROR_NAME_FOREIGN_KEY_VIOLATION {
				return ERROR_DB_FOREIGN_KEY_VIOLATION
			}
		}
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	rowDeleted, err := ret.RowsAffected()
	if err != nil {
		ctx.LogError("Error get rows affected = %v", err)
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	if rowDeleted == 0 {
		ctx.LogInfo("No row deleted in query: %s", query)
		return nil
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
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
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
		ctx.LogError("Error when get data = %#v, err = %s", data, err.Error())
		return err
	}

	primaryKeys, numPrimaryKeys := splitPrimaryKey(data)
	if numPrimaryKeys == 0 {
		ctx.LogError("Error not found primary key = %#v", data)
		return ERROR_NOT_FOUND_PRIMARY_KEY
	}

	for i, key := range primaryKeys {
		if i == 0 {
			query += fmt.Sprintf(" WHERE %s = $%d", key, i+1)
		} else {
			query += fmt.Sprintf(" AND %s = $%d", key, i+1)
		}
	}

	args, found := searchPrimaryKey(data)
	if !found {
		ctx.LogError("Error not found primary key = %#v", data)
		return ERROR_DB_ERROR
	}

	ctx.LogInfo("Select query = %v, args = %v", query, args)

	row := pgSession.QueryRowContext(ctx, query, args...)
	if err := row.Scan(params...); err != nil {
		ctx.LogError("Error select data = %#v, err = %v", data, err.Error())
		if err == sql.ErrNoRows {
			return ERROR_NOT_FOUND_IN_DB
		}
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return nil
}

/*
* ListAllInTable
* @params: ctx Context, data DataBaseObject
* @return []DataBaseObject, Error
* @description: select all data from table
 */
func ListAllInTable(ctx Context, data DataBaseObject) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

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
* ListTable
* @params: ctx Context, data DataBaseObject
* @return []DataBaseObject, Error
* @description: select all data from table
* @note: this function is used for paging
 */
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
* SelectListWithWhereQuery
* @params: ctx Context, data DataBaseObject, whereQuery string
* @return []DataBaseObject, Error
* @description: select list of data by where query
 */
func SelectListWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	query += " WHERE " + whereQuery

	ctx.LogInfo("Select query = %s", query)
	rows, errQuery := pgSession.QueryContext(ctx, query)
	if errQuery != nil {
		ctx.LogError("Error select query: %s | err = %s", query, errQuery.Error())
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
* SelectListWithAppendingQuery
* @params: ctx Context, data DataBaseObject, appendingQuery string
* @return []DataBaseObject, Error
* @description: select list of data by appending query
 */
func SelectListWithAppendingQuery(ctx Context, data DataBaseObject, appendingQuery string) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	query += " " + appendingQuery

	ctx.LogInfo("Select query = %s", query)
	rows, errQuery := pgSession.QueryContext(ctx, query)
	if errQuery != nil {
		ctx.LogError("Error select %s, err = %s", query, errQuery.Error())
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
* CountRecordInTable
* @params: ctx Context, data DataBaseObject
* @return int64, Error
* @description: count record in table
 */
func CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", data.GetTableName())

	row := pgSession.QueryRowContext(ctx, query)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		ctx.LogError("Error count record in table %s, err = %s", data.GetTableName(), err.Error())
		return 0, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return count, nil
}

/*
* CountRecordInTableWithWhere
* @params: ctx Context, data DataBaseObject, whereQuery string
* @return int64, Error
* @description: count record in table with where query
 */
func CountRecordInTableWithWhere(ctx Context, data DataBaseObject, whereQuery string) (int64, Error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", data.GetTableName(), whereQuery)
	ctx.LogInfo("Count record in table with where query: %s", query)
	row := pgSession.QueryRowContext(ctx, query)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		ctx.LogError("Error count record in table %s, err = %s", data.GetTableName(), err.Error())
		return 0, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return count, nil
}
