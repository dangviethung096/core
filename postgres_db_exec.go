package core

import (
	"database/sql"
	"reflect"

	"gorm.io/gorm"
)

type postgresSession struct {
	*gorm.DB
}

func (session postgresSession) InsertDataToDB(ctx Context, data DataBaseObject) Error {
	// Check data is a pointer and not nil
	if data == nil {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	ctx.LogInfo("Insert data table = %s, data = %#v", data.TableName(), data)
	if err := session.Create(data).Error; err != nil {
		ctx.LogError("Error insert data = %#v, err = %v", data, err)
		return ERROR_INSERT_TO_DB_FAIL
	}

	return nil
}

func (session postgresSession) DeleteDataFromDBByID(ctx Context, data DataBaseObject) Error {
	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	if data == nil {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}
	ctx.LogInfo("Delete data table = %s", data.TableName())
	if err := session.Delete(data).Error; err != nil {
		ctx.LogError("Error delete data = %#v, err = %v", data, err)
		return ERROR_DELETE_FROM_DB_FAIL
	}

	return nil
}

func (session postgresSession) DeleteDataFromDBWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string, args ...any) Error {
	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	if data == nil {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}
	ctx.LogInfo("Delete data table = %s, conditions = %v", data.TableName(), whereQuery)
	if err := session.Where(whereQuery, args...).Delete(data).Error; err != nil {
		ctx.LogError("Error delete data = %#v, err = %v", data, err)
		return ERROR_DELETE_FROM_DB_FAIL
	}

	return nil
}

func (session postgresSession) UpdateDataToDB(ctx Context, data DataBaseObject) Error {
	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	if data == nil {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}
	ctx.LogInfo("Update data table = %s, data = %#v", data.TableName(), data)
	if err := session.Save(data).Error; err != nil {
		ctx.LogError("Error update data = %#v, err = %s", data, err.Error())
		return ERROR_UPDATE_TO_DB_FAIL
	}

	return nil
}

func (session postgresSession) SelectListByFields(ctx Context, data DataBaseObject, whereQuery string, args ...interface{}) (any, Error) {
	if data == nil {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}

	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	// Get the type of the input data
	dataType := reflect.TypeOf(data)

	// If the type is a pointer, get the underlying struct type
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}

	// Create a slice of the same type
	sliceType := reflect.SliceOf(dataType)
	result := reflect.MakeSlice(sliceType, 0, 0).Interface()

	ctx.LogInfo("SelectListByFields, table = %s, conditions = %v", data.TableName(), whereQuery)

	if err := session.Where(whereQuery, args...).Find(&result).Error; err != nil {
		ctx.LogError("Error select list by fields = %#v, err = %s", data, err.Error())
		return nil, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return result, nil
}

func (session postgresSession) SelectListByFieldWithPaging(ctx Context, data DataBaseObject, limit int64, offset int64, whereQuery string, args ...any) (any, Error) {
	if data == nil {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}

	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	// Get the type of the input data
	dataType := reflect.TypeOf(data)

	// If the type is a pointer, get the underlying struct type
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}

	// Create a slice of the same type
	sliceType := reflect.SliceOf(dataType)
	result := reflect.MakeSlice(sliceType, 0, 0).Interface()

	ctx.LogInfo("SelectListByFieldWithPaging, table = %s, conditions = %v, limit = %d, offset = %d", data.TableName(), whereQuery, limit, offset)
	if err := session.Where(whereQuery, args...).Limit(int(limit)).Offset(int(offset)).Find(&result).Error; err != nil {
		ctx.LogError("Error select list by fields = %#v, err = %s", data, err.Error())
		return nil, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return result, nil
}

func (session postgresSession) CountRecordInTable(ctx Context, data DataBaseObject) (int64, Error) {
	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return 0, NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	if data == nil {
		return 0, NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}

	ctx.LogInfo("CountRecordInTable, table = %s", data.TableName())
	var count int64
	if err := session.Model(&data).Count(&count).Error; err != nil {
		ctx.LogError("Error count record in table = %#v, err = %s", data, err.Error())
		return 0, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return count, nil
}

func (session postgresSession) CountRecordInTableWithWhereQuery(ctx Context, data DataBaseObject, whereQuery string, args ...any) (int64, Error) {
	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return 0, NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	if data == nil {
		return 0, NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}

	ctx.LogInfo("CountRecordInTableWithWhereQuery, table = %s, conditions = %v", data.TableName(), whereQuery)
	var count int64
	if err := session.Model(&data).Where(whereQuery, args...).Count(&count).Error; err != nil {
		ctx.LogError("Error count record in table = %#v, err = %s", data, err.Error())
		return 0, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return count, nil
}

func (session postgresSession) SelectPaging(ctx Context, data DataBaseObject, orderQuery string, limit int64, offset int64) (any, Error) {
	if data == nil {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}

	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return nil, NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	// Get the type of the input data
	dataType := reflect.TypeOf(data)

	// If the type is a pointer, get the underlying struct type
	if dataType.Kind() == reflect.Ptr {
		dataType = dataType.Elem()
	}

	// Create a slice of the same type
	sliceType := reflect.SliceOf(dataType)
	result := reflect.MakeSlice(sliceType, 0, 0).Interface()

	ctx.LogInfo("SelectPaging, table = %s, limit = %d, offset = %d", data.TableName(), limit, offset)

	if err := session.Order(orderQuery).Limit(int(limit)).Offset(int(offset)).Find(&result).Error; err != nil {
		ctx.LogError("Error select list by fields = %#v, err = %s", data, err.Error())
		return nil, NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return result, nil
}

func (session postgresSession) Close() {
	sqlDB, err := session.DB.DB()
	if err != nil {
		return
	}
	sqlDB.Close()
}

func (session postgresSession) Connection() *gorm.DB {
	return session.DB
}

// Transaction executes the given function within a transaction
func (session postgresSession) Transaction(ctx Context, fn func(tx *gorm.DB) error) Error {
	err := session.DB.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
	if err != nil {
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return nil
}

func (session postgresSession) GetOriginConnection(ctx Context) *sql.DB {
	sqlDB, err := session.DB.DB()
	if err != nil {
		ctx.LogError("Error get origin connection = %s", err.Error())
		return nil
	}
	return sqlDB
}

func (session postgresSession) AutoMigrate(ctx Context, data DataBaseObject) Error {
	if err := session.DB.AutoMigrate(data); err != nil {
		ctx.LogError("Error auto migrate table = %#v, err = %s", data, err.Error())
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return nil
}

func (session postgresSession) SelectByID(ctx Context, data DataBaseObject) Error {
	if data == nil {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}

	// Check data is a pointer and not nil
	if reflect.TypeOf(data).Kind() != reflect.Ptr {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is not a pointer")
	}

	if err := session.First(data).Error; err != nil {
		ctx.LogError("Error select by id = %#v, err = %s", data, err.Error())
		if err == gorm.ErrRecordNotFound {
			return ERROR_NOT_FOUND_IN_DB
		}
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return nil
}

func (session postgresSession) SelectOneByField(ctx Context, data DataBaseObject, whereQuery string, args ...any) Error {
	if data == nil {
		return NewError(ERROR_CODE_FROM_DATABASE, "data is nil")
	}

	if err := session.Where(whereQuery, args...).First(data).Error; err != nil {
		ctx.LogError("Error select one by field = %#v, err = %s", data, err.Error())
		if err == gorm.ErrRecordNotFound {
			ctx.LogInfo("No record found with query = %v, args = %v", whereQuery, args)
			return ERROR_NOT_FOUND_IN_DB
		}
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return nil
}

func (session postgresSession) DeleteDataFromDBByWhereQuery(ctx Context, data DataBaseObject, whereQuery string, args ...any) Error {
	ctx.LogInfo("DeleteDataFromDBByWhereQuery, table = %s, conditions = %v, args = %v", data.TableName(), whereQuery, args)
	if err := session.Where(whereQuery, args...).Delete(data).Error; err != nil {
		ctx.LogError("Error delete data by where query = %#v, err = %s", data, err.Error())
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}
	return nil
}
