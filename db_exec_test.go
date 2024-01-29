package core

import (
	"testing"
)

/*
DROP TABLE IF EXISTS test_accounts;

CREATE TABLE test_accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    age INT NOT NULL,
);
*/

type Account struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func (a Account) GetTableName() string {
	return "test_accounts"
}

func (a Account) GetPrimaryKey() string {
	return "id"
}

var account1 = Account{
	Id:   1,
	Name: "Hung",
	Age:  11,
}

var account2 = Account{
	Id:   2,
	Name: "Dat",
	Age:  18,
}

var account3 = Account{
	Id:   3,
	Name: "Hoang",
	Age:  20,
}

func insertAccount(ctx *Context) {
	SaveDataToDB(ctx, &account1)
	SaveDataToDB(ctx, &account2)
	SaveDataToDB(ctx, &account3)
}

func deleteAccount(ctx *Context) {
	DeleteDataInDB(ctx, &account1)
	DeleteDataInDB(ctx, &account2)
	DeleteDataInDB(ctx, &account3)
}

func TestSelectListByNameField_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	insertAccount(ctx)
	defer deleteAccount(ctx)
	var account Account
	result, err := SelectListByField(ctx, &account, "name", "Hung")
	if err != nil {
		t.Errorf("TestSelectListByField_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", result)
}

func TestSelectListByAgeField_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	insertAccount(ctx)
	defer deleteAccount(ctx)
	var account Account
	result, err := SelectListByField(ctx, &account, "age", 11)
	if err != nil {
		t.Errorf("TestSelectListByField_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", result)
}

func TestSelectByField_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	insertAccount(ctx)
	defer deleteAccount(ctx)
	var account Account
	err := SelectByField(ctx, &account, "name", "Hung")
	if err != nil {
		t.Errorf("TestSelectByField_ReturnSuccess: %v", err)
	}

	if account.Name != "Hung" {
		t.Errorf("TestSelectByField_ReturnSuccess: expected name to be 'Hung', got '%s'", account.Name)
	}
}

func TestSelectByField_InvalidField_ReturnError(t *testing.T) {
	ctx := GetContextForTest()
	var account Account
	insertAccount(ctx)
	defer deleteAccount(ctx)
	err := SelectByField(ctx, &account, "invalid_field", "Hung")
	if err == nil {
		t.Errorf("TestSelectByField_InvalidField_ReturnError: expected an error, got nil")
	}
}

func TestSelectByField_NotFound_ReturnError(t *testing.T) {
	ctx := GetContextForTest()
	insertAccount(ctx)
	defer deleteAccount(ctx)
	var account Account
	err := SelectByField(ctx, &account, "name", "Nonexistent")
	if err == nil {
		t.Errorf("TestSelectByField_NotFound_ReturnError: expected an error, got nil")
	}
}

func TestSelectListByFieldsWithCustomOperator_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	var account Account
	insertAccount(ctx)
	defer deleteAccount(ctx)

	values, err := SelectListByFieldsWithCustomOperator(ctx, &account, DBWhere{
		FieldName: "age",
		Operator:  ">",
		Value:     12,
	})
	if err != nil {
		t.Errorf("TestSelectListByFieldsWithCustomOperator_ReturnSuccess: get error: %s\n", err.Error())
		return
	}

	if len(values) != 2 {
		t.Errorf("TestSelectListByFieldsWithCustomOperator_ReturnSuccess: wrong return values: %#v\n", values)
		return
	}

	t.Logf("Values: %#v\n", values)
}
