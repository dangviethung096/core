package core

import (
	"database/sql"

	"github.com/lib/pq"
)

type oracleSession struct {
	*sql.DB
}

func (session oracleSession) SaveDataToDB(ctx Context, data DataBaseObject) Error {
	query, args, insertError := GetInsertQuery(data)
	if insertError != nil {
		ctx.LogError("Error when get insert data = %#v, err = %s", data, insertError.Error())
		return insertError
	}

	ctx.LogInfo("Insert query = %v, args = %v", query, args)
	if _, err := session.ExecContext(ctx, query, args...); err != nil {
		ctx.LogError("Error insert data = %#v, err = %v", data, err)
		return ERROR_INSERT_TO_DB_FAIL
	}

	return nil
}

func (session oracleSession) SaveDataToDBWithoutPrimaryKey(ctx Context, data DataBaseObject) Error {
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

func (session oracleSession) DeleteDataInDB(ctx Context, data DataBaseObject) Error {
	return nil
}

func (session oracleSession) DeleteDataWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) Error {
	return nil
}

func (session oracleSession) UpdateDataInDB(ctx Context, data DataBaseObject) Error {
	return nil
}

func (session oracleSession) SelectById(ctx Context, data DataBaseObject) Error {
	return nil
}

func (session oracleSession) ListAllInTable(ctx Context, data DataBaseObject) (any, Error) {
	return nil, nil
}

func (session oracleSession) SelectListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}) (any, Error) {
	return nil, nil
}

func (session oracleSession) SelectListWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) (any, Error) {
	return nil, nil
}

func (session oracleSession) ListPagingTable(ctx Context, data DataBaseObject, limit int64, offset int64) (any, Error) {
	return nil, nil
}

func (session oracleSession) SelectPagingListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64) (any, Error) {
	return nil, nil
}

func (session oracleSession) CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error) {
	return 0, nil
}

func (session oracleSession) CountRecordInTableWithWhere(ctx Context, data DataBaseObject, whereQuery string) (int64, Error) {
	return 0, nil
}
