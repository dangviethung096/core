package core

import "reflect"

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
