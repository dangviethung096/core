package core

import "reflect"

/*
* Save data to database
* @param data interface{} Data to save
* @return Error
 */
func SaveDataToDB[T DataBaseObject](ctx Context, data T) Error {
	return mainDbSession.SaveDataToDB(ctx, data)
}

/*
* Save data to database without primary key
* primary key will be auto increment in database
* @param data interface{} Data to save
* @return Error
 */
func SaveDataToDBWithoutPrimaryKey[T DataBaseObject](ctx Context, data T) Error {
	return mainDbSession.SaveDataToDBWithoutPrimaryKey(ctx, data)
}

/*
* Delete data in database
* @param data interface{} Data to delete
* @return Error
 */
func DeleteDataInDB[T DataBaseObject](ctx Context, data T) Error {
	return mainDbSession.DeleteDataInDB(ctx, data)
}

func DeleteDataWithWhereQuery[T DataBaseObject](ctx Context, data T, whereQuery string) Error {
	return mainDbSession.DeleteDataWithWhereQuery(ctx, data, whereQuery)
}

/*
* Update data in database
* @param data interface{} Data to update
* @return Error
 */
func UpdateDataInDB[T DataBaseObject](ctx Context, data T) Error {
	return mainDbSession.UpdateDataInDB(ctx, data)
}

/*
* Select data from database by primary key
* @param data interface{} Data to select
* @return Error
 */
func SelectById(ctx Context, data DataBaseObject) Error {
	return mainDbSession.SelectById(ctx, data)
}

/*
* ListAllInTable
* @params: ctx Context, data DataBaseObject
* @return []DataBaseObject, Error
* @description: select all data from table
 */
func ListAllInTable(ctx Context, data DataBaseObject) (any, Error) {
	return mainDbSession.ListAllInTable(ctx, data)
}

/*
* ListTable
* @params: ctx Context, data DataBaseObject
* @return []DataBaseObject, Error
* @description: select all data from table
* @note: this function is used for paging
 */
func ListPagingTable(ctx Context, data DataBaseObject, limit int64, offset int64) (any, Error) {
	return mainDbSession.ListPagingTable(ctx, data, limit, offset)
}

/*
* Select data from database by field: fieldName and fieldValue is passed in parameter
* @return Error
 */
func SelectByField(ctx Context, data DataBaseObject, fieldName string, fieldValue any) Error {
	result, err := mainDbSession.SelectPagingListByFields(ctx, data, map[string]interface{}{
		fieldName: fieldValue,
	}, 1, 0)
	if err != nil {
		return err
	}

	if len(result.([]*DataBaseObject)) != 0 {
		// Copy value of result to data
		reflect.ValueOf(data).Elem().Set(reflect.ValueOf(result.([]*DataBaseObject)[0]))
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
	return mainDbSession.SelectListByFields(ctx, data, map[string]interface{}{
		fieldName: fieldValue,
	})
}

func SelectListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}) (any, Error) {
	return mainDbSession.SelectListByFields(ctx, data, mapArgs)
}

/*
* SelectListWithWhereQuery
* @params: ctx Context, data DataBaseObject, whereQuery string
* @return []DataBaseObject, Error
* @description: select list of data by where query
 */
func SelectListWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string) (any, Error) {
	return mainDbSession.SelectListWithWhereQuery(ctx, data, whereQuery)
}

/*
* SelectPagingListByFields
* @params: ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64
* @return []DataBaseObject, Error
* @description: select list of data by args with limit and offset
* @note: this function is used for paging
 */
func SelectPagingListByFields(ctx Context, data DataBaseObject, mapArgs map[string]interface{}, limit int64, offset int64) (any, Error) {
	return mainDbSession.SelectPagingListByFields(ctx, data, mapArgs, limit, offset)
}

/*
* CountRecordInTable
* @params: ctx Context, data DataBaseObject
* @return int64, Error
* @description: count record in table
 */
func CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error) {
	return mainDbSession.CountRecordInTable(ctx, data)
}

/*
* CountRecordInTableWithWhere
* @params: ctx Context, data DataBaseObject, whereQuery string
* @return int64, Error
* @description: count record in table with where query
 */
func CountRecordInTableWithWhere(ctx Context, data DataBaseObject, whereQuery string) (int64, Error) {
	return mainDbSession.CountRecordInTableWithWhere(ctx, data, whereQuery)
}
