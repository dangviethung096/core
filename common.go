package core

import "reflect"

/*
* getTypeOfPointer: get type of pointer
* @params: data any
* @return: reflect.Type
 */
func getTypeOfPointer(p any) (reflect.Type, Error) {
	if p == nil {
		return nil, ERROR_NIL_PARAM
	}

	t := reflect.TypeOf(p)
	// Check if model is a struct
	if !(t.Kind() == reflect.Pointer) {
		return nil, ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT
	}
	t = t.Elem()
	if !(t.Kind() == reflect.Struct) {
		return nil, ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT
	}

	return t, nil
}

/*
* paramIsPointerOfStruct: check if param is pointer of struct
* @params: data any
 */
func paramIsPointerOfStruct(p any) Error {
	if p == nil {
		return ERROR_NIL_PARAM
	}

	t := reflect.TypeOf(p)
	// Check if model is a struct
	if !(t.Kind() == reflect.Pointer) {
		return ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT
	}
	t = t.Elem()
	if !(t.Kind() == reflect.Struct) {
		return ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT
	}
	return nil
}
