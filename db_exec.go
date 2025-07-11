package core

/*
* Save data to database
* @param data any Data to save
* @return Error
 */
func InsertDataToDB(ctx Context, data DataBaseObject) Error {
	return mainDbSession.InsertDataToDB(ctx, data)
}

/*
* Delete data in database
* @param data any Data to delete
* @return Error
 */
func DeleteDataFromDB(ctx Context, data DataBaseObject) Error {
	return mainDbSession.DeleteDataFromDBByID(ctx, data)
}

/*
* Update data in database
* @param data any Data to update
* @return Error
 */
func UpdateDataToDB(ctx Context, data DataBaseObject) Error {
	return mainDbSession.UpdateDataToDB(ctx, data)
}

func SelectListByFields(ctx Context, data DataBaseObject, whereQuery string, args ...any) (any, Error) {
	return mainDbSession.SelectListByFields(ctx, data, whereQuery, args...)
}

func SelectListByFieldWithPaging(ctx Context, data DataBaseObject, orderQuery string, limit int64, offset int64, whereQuery string, args ...any) (any, Error) {
	return mainDbSession.SelectListByFieldWithPaging(ctx, data, orderQuery, limit, offset, whereQuery, args...)
}

func SelectPaging(ctx Context, data DataBaseObject, orderQuery string, limit int64, offset int64) (any, Error) {
	return mainDbSession.SelectPaging(ctx, data, orderQuery, limit, offset)
}

func CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error) {
	return mainDbSession.CountRecordInTable(ctx, data)
}

func SelectByID(ctx Context, data DataBaseObject) Error {
	return mainDbSession.SelectByID(ctx, data)
}

func CountRecordInTableWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string, args ...any) (int64, Error) {
	return mainDbSession.CountRecordInTableWithWhereQuery(ctx, data, whereQuery, args...)
}

func SelectOneByField(ctx Context, data DataBaseObject, whereQuery string, args ...any) Error {
	return mainDbSession.SelectOneByField(ctx, data, whereQuery, args...)
}

func DeleteDataFromDBByWhereQuery(ctx Context, data DataBaseObject, whereQuery string, args ...any) Error {
	return mainDbSession.DeleteDataFromDBByWhereQuery(ctx, data, whereQuery, args...)
}
