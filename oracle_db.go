package core

import "database/sql"

type oracleSession struct {
	*sql.DB
}

func (session oracleSession) SaveDataToDB(ctx Context, data DataBaseObject) Error {
	return nil
}

func (session oracleSession) SaveDataToDBWithoutPrimaryKey(ctx Context, data DataBaseObject) Error {
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
