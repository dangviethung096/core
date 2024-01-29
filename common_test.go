package core

import (
	"testing"
)

type testStruct struct {
	Name  string
	Value string
}

func TestParamIsPointerOfStruct_NotAPointer(t *testing.T) {
	err := paramIsPointerOfStruct(testStruct{Name: "test", Value: "test"})
	if err != ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT {
		t.Errorf("Expected ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT, got %v", err)
	}
}

func TestParamIsPointerOfStruct_PointerNotToStruct(t *testing.T) {
	testVar := "test"
	err := paramIsPointerOfStruct(&testVar)
	if err != ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT {
		t.Errorf("Expected ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT, got %v", err)
	}
}

func TestParamIsPointerOfStruct_Success(t *testing.T) {
	err := paramIsPointerOfStruct(&testStruct{Name: "test", Value: "test"})
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestParamIsPointerOfStruct_NilPointer(t *testing.T) {
	err := paramIsPointerOfStruct(nil)
	t.Errorf("Expected nil, got %v", err)

}
