package core

import (
	"fmt"
	"testing"
)

/*
DROP TABLE IF EXISTS test_accounts;

CREATE TABLE test_accounts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    age INT NOT NULL
);


DROP TABLE IF EXISTS test_account_many_key;

CREATE TABLE test_account_many_key (
    id INT NOT NULL,
    name TEXT NOT NULL,
    age INT NOT NULL,
	note TEXT,
	PRIMARY KEY (id, name)
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

type AccountManyKey struct {
	Id        int    `db:"id"`
	Name      string `db:"name"`
	Age       int    `db:"age"`
	Note      string `db:"note"`
	NoField   string
	NoFiled02 string
}

func (a AccountManyKey) GetTableName() string {
	return "test_account_many_key"
}

func (a AccountManyKey) GetPrimaryKey() string {
	return "id,name"
}

var accountMany01 = AccountManyKey{
	Id:   1,
	Name: "Hung",
	Age:  11,
	Note: "Note 1",
}

var accountMany02 = AccountManyKey{
	Id:   2,
	Name: "Dat",
	Age:  18,
	Note: "Note 2",
}

var accountMany03 = AccountManyKey{
	Id:   3,
	Name: "Hoang",
	Age:  20,
	Note: "Note 3",
}

func insertAccount(ctx Context) {
	SaveDataToDB(ctx, &account1)
	SaveDataToDB(ctx, &account2)
	SaveDataToDB(ctx, &account3)
}

func insertAccountWithManyKey(ctx Context) {
	SaveDataToDB(ctx, &accountMany01)
	SaveDataToDB(ctx, &accountMany02)
	SaveDataToDB(ctx, &accountMany03)
}

func deleteAccount(ctx Context) {
	DeleteDataInDB(ctx, &account1)
	DeleteDataInDB(ctx, &account2)
	DeleteDataInDB(ctx, &account3)
}

func deleteAccountWithManyKey(ctx Context) {
	DeleteDataInDB(ctx, &accountMany01)
	DeleteDataInDB(ctx, &accountMany02)
	DeleteDataInDB(ctx, &accountMany03)
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

	if len(values.([]Account)) != 2 {
		t.Errorf("TestSelectListByFieldsWithCustomOperator_ReturnSuccess: wrong return values: %#v\n", values)
		return
	}

	t.Logf("Values: %#v\n", values)
}

func TestSelectListWithWhereQuery_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	var account Account
	insertAccount(ctx)
	defer deleteAccount(ctx)

	whereQuery := fmt.Sprintf("age > %d", 18)
	result, err := SelectListWithWhereQuery(ctx, &account, whereQuery)
	if err != nil {
		t.Errorf("TestSelectListWithWhereQuery_ReturnSuccess: get error: %s\n", err.Error())
		return
	}

	t.Logf("Result: %#v\n", result)
}

func TestSelectById_ReturnSuccessWithManyKey(t *testing.T) {
	ctx := GetContextForTest()
	insertAccountWithManyKey(ctx)
	defer deleteAccountWithManyKey(ctx)

	account := AccountManyKey{
		Id:   1,
		Name: "Hung",
	}
	err := SelectById(ctx, &account)
	if err != nil {
		t.Errorf("TestSelectByPrimaryKey_ReturnSuccess: get error: %s\n", err.Error())
		return
	}

	if account.Age != accountMany01.Age {
		t.Errorf("TestSelectByPrimaryKey_ReturnSuccess: wrong return values: %#v\n", account)
		return
	}

	t.Logf("Account: %#v\n", account)
}

func TestUpdateDataInDB_ReturnSuccessWithManyKey(t *testing.T) {
	ctx := GetContextForTest()
	insertAccountWithManyKey(ctx)
	defer deleteAccountWithManyKey(ctx)

	account := AccountManyKey{
		Id:   1,
		Name: "Hung",
		Age:  23,
		Note: "TestNote",
	}

	err := UpdateDataInDB(ctx, &account)
	if err != nil {
		t.Errorf("TestUpdateDataInDB_ReturnSuccessWithManyKey: get error: %s\n", err.Error())
		return
	}

	// Get account from db
	acc := AccountManyKey{
		Id:   1,
		Name: "Hung",
	}

	err = SelectById(ctx, &acc)
	if err != nil {
		t.Errorf("TestUpdateDataInDB_ReturnSuccessWithManyKey: get error: %s\n", err.Error())
		return
	}

	if acc.Age != account.Age || acc.Note != account.Note {
		t.Errorf("TestUpdateDataInDB_ReturnSuccessWithManyKey: wrong return values: %#v\n", acc)
		return
	}

	t.Logf("Account: %#v\n", acc)
}

func TestDeleteDataInDB_ReturnSuccessWithManyKey(t *testing.T) {
	ctx := GetContextForTest()
	insertAccountWithManyKey(ctx)
	defer deleteAccountWithManyKey(ctx)

	account := AccountManyKey{
		Id:   1,
		Name: "Hung",
	}

	err := DeleteDataInDB(ctx, &account)
	if err != nil {
		t.Errorf("TestDeleteDataInDB_ReturnSuccessWithManyKey: get error: %s\n", err.Error())
		return
	}

	err = SelectById(ctx, &account)
	if err == nil {
		t.Errorf("TestDeleteDataInDB_ReturnSuccessWithManyKey: expected an error, got nil")
		return
	}

	t.Logf("Delete Account: %#v\n", account)
}

func TestSelectListWithAppendingQuery_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	var account Account
	insertAccount(ctx)
	defer deleteAccount(ctx)

	result, err := SelectListWithAppendingQuery(ctx, &account, "LIMIT 1")
	if err != nil {
		t.Errorf("TestSelectListWithAppendingQuery_ReturnSuccess: get error: %s\n", err.Error())
		return
	}

	if len(result.([]Account)) != 1 {
		t.Errorf("TestSelectListWithAppendingQuery_ReturnSuccess: wrong return values: %#v\n", result)
		return
	}

	t.Logf("Result: %#v\n", result)
}
