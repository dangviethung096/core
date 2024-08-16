package core

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/lib/pq"
)

type postgresSession struct {
	*sql.DB
}

func (session postgresSession) SaveDataToDB(ctx Context, data DataBaseObject) Error {
	query, args, insertError := GetInsertQuery(data)
	if insertError != nil {
		ctx.LogError("Error when get insert data = %#v, err = %s", data, insertError.Error())
		return insertError
	}

	ctx.LogInfo("Insert query = %v, args = %v", query, args)
	if _, err := session.ExecContext(ctx, query, args...); err != nil {
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

func (session postgresSession) SaveDataToDBWithoutPrimaryKey(ctx Context, data DataBaseObject) Error {
	query, args, pkAddress, insertError := GetInsertQueryWithoutPrimaryKey(data)
	if insertError != nil {
		ctx.LogError("Error when get insert data = %#v, err = %s", data, insertError.Error())
		return insertError
	}

	ctx.LogInfo("Insert query = %v, args = %v", query, args)
	row := session.QueryRowContext(ctx, query, args...)

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

func (session postgresSession) DeleteDataInDB(ctx Context, data DataBaseObject) Error {
	query, args, deleteError := GetDeleteQuery(data)
	if deleteError != nil {
		ctx.LogError("Error when get delete data = %#v, err = %s", data, deleteError.Error())
		return deleteError
	}

	ctx.LogInfo("Delete query = %v, args = %v", query, args)
	if _, err := session.ExecContext(ctx, query, args...); err != nil {
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

func (session postgresSession) DeleteDataWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) Error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", data.GetTableName(), whereQuery)

	ctx.LogInfo("Delete query = %v", query)
	ret, err := session.ExecContext(ctx, query)
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

func (session postgresSession) UpdateDataInDB(ctx Context, data DataBaseObject) Error {
	query, args, updateError := GetUpdateQuery(data)
	if updateError != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, updateError.Error())
		return updateError
	}

	ctx.LogInfo("Update query = %v, args = %v", query, args)
	if _, err := session.ExecContext(ctx, query, args...); err != nil {
		ctx.LogError("Error update data = %#v, err = %s", data, err.Error())
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return nil
}

func (session postgresSession) SelectById(ctx Context, data DataBaseObject) Error {
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

	row := session.QueryRowContext(ctx, query, args...)
	if err := row.Scan(params...); err != nil {
		ctx.LogError("Error select data = %#v, err = %v", data, err.Error())
		if err == sql.ErrNoRows {
			return ERROR_NOT_FOUND_IN_DB
		}
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return nil
}

func (session postgresSession) ListAllInTable(ctx Context, data DataBaseObject) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	ctx.LogInfo("Select query = %v", query)
	rows, errQuery := session.QueryContext(ctx, query)
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
func (session postgresSession) ListPagingTable(ctx Context, data DataBaseObject, limit int64, offset int64) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	ctx.LogInfo("Select query = %v", query)
	rows, errQuery := session.QueryContext(ctx, query)
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

func (session postgresSession) SelectByField(ctx Context, data DataBaseObject, fieldName string, fieldValue any) Error {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return err
	}

	query += fmt.Sprintf(" WHERE %s = $1", fieldName)

	ctx.LogInfo("Select query = %v, args = %v", query, fieldValue)
	row := session.QueryRowContext(ctx, query, fieldValue)
	if err := row.Scan(params...); err != nil {
		ctx.LogError("Error select data = %#v, err = %s", data, err.Error())
		if err == sql.ErrNoRows {
			return ERROR_NOT_FOUND_IN_DB
		}
		return ERROR_DB_ERROR
	}

	return nil
}

func (session postgresSession) SelectListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}) (any, Error) {
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
	rows, errQuery := session.QueryContext(ctx, query, args...)
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

func (session postgresSession) SelectListWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) (any, Error) {
	query, params, err := GetSelectQuery(data)
	if err != nil {
		ctx.LogError("Error when get update data = %#v, err = %s", data, err.Error())
		return nil, err
	}

	query += " WHERE " + whereQuery

	ctx.LogInfo("Select query = %s", query)
	rows, errQuery := session.QueryContext(ctx, query)
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

func (session postgresSession) SelectPagingListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64) (any, Error) {
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
	rows, errQuery := session.QueryContext(ctx, query, args...)
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

func (session postgresSession) CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", data.GetTableName())

	row := session.QueryRowContext(ctx, query)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		ctx.LogError("Error count record in table %s, err = %s", data.GetTableName(), err.Error())
		return 0, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return count, nil
}

func (session postgresSession) CountRecordInTableWithWhere(ctx Context, data DataBaseObject, whereQuery string) (int64, Error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", data.GetTableName(), whereQuery)
	ctx.LogInfo("Count record in table with where query: %s", query)
	row := session.QueryRowContext(ctx, query)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		ctx.LogError("Error count record in table %s, err = %s", data.GetTableName(), err.Error())
		return 0, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return count, nil
}
