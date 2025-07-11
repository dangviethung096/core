package core

import (
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
	Id   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name;not null"`
	Age  int    `gorm:"column:age;not null"`
}

func (a Account) TableName() string {
	return "test_accounts"
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
	Id        int    `gorm:"column:id"`
	Name      string `gorm:"column:name;primaryKey"`
	Age       int    `gorm:"column:age;primaryKey"`
	Note      string `gorm:"column:note"`
	NoField   string `gorm:"-"` // Ignored by GORM
	NoField02 string `gorm:"-"` // Ignored by GORM
}

// TableName specifies the table name for GORM
func (AccountManyKey) TableName() string {
	return "test_account_many_key"
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
	InsertDataToDB(ctx, &account1)
	InsertDataToDB(ctx, &account2)
	InsertDataToDB(ctx, &account3)
}

func insertAccountWithManyKey(ctx Context) {
	InsertDataToDB(ctx, &accountMany01)
	InsertDataToDB(ctx, &accountMany02)
	InsertDataToDB(ctx, &accountMany03)
}

func deleteAccount(ctx Context) {
	DeleteDataFromDB(ctx, account1)
	DeleteDataFromDB(ctx, account2)
	DeleteDataFromDB(ctx, account3)
}

func deleteAccountWithManyKey(ctx Context) {
	DeleteDataFromDB(ctx, accountMany01)
	DeleteDataFromDB(ctx, accountMany02)
	DeleteDataFromDB(ctx, accountMany03)
}

// migrateTables performs database migrations
func migrateTables(ctx Context) Error {
	// Drop table if exists
	DBSession().Connection().Exec("DROP TABLE IF EXISTS test_accounts")
	DBSession().Connection().Exec("DROP TABLE IF EXISTS test_account_many_key")

	// Migrate Account table
	if err := DBSession().AutoMigrate(ctx, &Account{}); err != nil {
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	// Migrate AccountManyKey table
	if err := DBSession().AutoMigrate(ctx, &AccountManyKey{}); err != nil {
		return NewError(ERROR_CODE_FROM_DATABASE, err.Error())
	}

	return nil
}

func TestSelectListByNameField_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccount(ctx)
	defer deleteAccount(ctx)
	var account Account
	accounts, err := SelectListByFields(ctx, account, "name = ?", "Hung")
	if err != nil {
		t.Errorf("TestSelectListByField_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", accounts)
}

func TestSelectListByManyKeyField_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccountWithManyKey(ctx)
	defer deleteAccountWithManyKey(ctx)
	var account AccountManyKey
	accounts, err := SelectListByFields(ctx, account, "id = ?", 1)
	if err != nil {
		t.Errorf("TestSelectListByManyKeyField_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", accounts)
}

func TestCreateData_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()

	account := Account{
		Id:   1,
		Name: "Hung",
		Age:  11,
	}
	err := InsertDataToDB(ctx, &account)
	defer DeleteDataFromDB(ctx, account)
	if err != nil {
		t.Errorf("TestCreateData_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", account)
}

func TestCreateDataWithManyKey_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()

	account := AccountManyKey{
		Id:   1,
		Name: "Hung",
		Age:  11,
		Note: "Note 1",
	}
	err := InsertDataToDB(ctx, &account)
	defer DeleteDataFromDB(ctx, account)
	if err != nil {
		t.Errorf("TestCreateDataWithManyKey_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", account)
}

func TestUpdateData_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccount(ctx)
	defer deleteAccount(ctx)

	account := Account{
		Id:   1,
		Name: "Hung",
		Age:  11,
	}
	err := UpdateDataToDB(ctx, &account)
	if err != nil {
		t.Errorf("TestUpdateData_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", account)
}

func TestSelectByID_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	insertAccount(ctx)
	defer deleteAccount(ctx)

	account := Account{
		Id: 1,
	}
	err := SelectByID(ctx, &account)
	if err != nil {
		t.Errorf("TestSelectByID_ReturnSuccess: %v", err)
		return
	}

	t.Logf("Result: %#v", account)
}

func TestSelectByManyKey_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccountWithManyKey(ctx)
	defer deleteAccountWithManyKey(ctx)

	account := AccountManyKey{
		Name: "Hung",
		Age:  11,
	}

	err := SelectByID(ctx, &account)
	if err != nil {
		t.Errorf("TestSelectByManyKey_ReturnSuccess: %v", err)
		return
	}

	t.Logf("Result: %#v", account)
}

func TestCountRecordInTable_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccount(ctx)
	defer deleteAccount(ctx)

	count, err := CountRecordInTable(ctx, &Account{})
	if err != nil {
		t.Errorf("TestCountRecordInTable_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %d", count)
}

func TestCountRecordInTableWithWhereQuery_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccount(ctx)
	defer deleteAccount(ctx)

	count, err := CountRecordInTableWithWhereQuery(ctx, &Account{}, "name = ?", "Hung")
	if err != nil {
		t.Errorf("TestCountRecordInTableWithWhereQuery_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %d", count)
}

func TestSelectPaging_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccount(ctx)
	defer deleteAccount(ctx)

	accounts, err := SelectPaging(ctx, &Account{}, "name = ?", 1, 1)
	if err != nil {
		t.Errorf("TestSelectPaging_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", accounts)
}

func TestSelectListByFieldWithPaging_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccount(ctx)
	defer deleteAccount(ctx)

	accounts, err := SelectListByFieldWithPaging(ctx, &Account{}, "name asc", 1, 0, "name = ?", "Hung")
	if err != nil {
		t.Errorf("TestSelectListByFieldWithPaging_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", accounts)
}

func TestSelectListByFields_ReturnSuccess(t *testing.T) {
	ctx := GetContextForTest()
	migrateTables(ctx)
	insertAccount(ctx)
	defer deleteAccount(ctx)

	accounts, err := SelectListByFields(ctx, &Account{}, "name = ?", "Hung")
	if err != nil {
		t.Errorf("TestSelectListByFields_ReturnSuccess: %v", err)
	}

	t.Logf("Result: %#v", accounts)
}
