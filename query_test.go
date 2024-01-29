package core

import (
	"reflect"
	"testing"
)

type UserTest struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

func (u UserTest) GetTableName() string {
	return "users"
}

func (u UserTest) GetPrimaryKey() string {
	return "id"
}

type UserTestNotFoundPrimaryKey struct {
	Id   int    `db:"not_pk"`
	Name string `db:"name"`
}

func (u UserTestNotFoundPrimaryKey) GetTableName() string {
	return "users"
}

func (u UserTestNotFoundPrimaryKey) GetPrimaryKey() string {
	return "id"
}

type UserInvalid struct {
}

func (u UserInvalid) GetTableName() string {
	return "users"
}

func (u UserInvalid) GetPrimaryKey() string {
	return "id"
}

func TestGetSelectQuery_Success(t *testing.T) {
	user := &UserTest{Id: 1, Name: "John"}

	wantQuery := "SELECT id, name FROM users"
	wantArgs := []interface{}{
		&user.Id,
		&user.Name,
	}

	gotQuery, gotArgs, err := GetSelectQuery(user)

	if err != nil {
		t.Errorf("GetSelectQuery() error = %v, wantErr %v", err, false)
	}

	if gotQuery != wantQuery {
		t.Errorf("GetSelectQuery() query = %v, want %v", gotQuery, wantQuery)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("GetSelectQuery() args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestGetSelectQuery_ErrorIfModelIsNotPointerToStruct(t *testing.T) {
	user := UserTest{Id: 1, Name: "John"}

	_, _, err := GetSelectQuery(user)

	if err != ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT {
		t.Errorf("GetSelectQuery() error = %v, wantErr %v", err, ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT)
	}
}

func TestGetSelectQuery_ErrorIfModelHasNoFields(t *testing.T) {
	user := UserInvalid{}

	_, _, err := GetSelectQuery(&user)

	if err != ERROR_MODEL_HAVE_NO_FIELD {
		t.Errorf("GetSelectQuery() error = %v, wantErr %v", err, ERROR_MODEL_HAVE_NO_FIELD)
	}
}

func TestGetInsertQuery_Success(t *testing.T) {
	user := UserTest{Id: 1, Name: "John"}

	wantQuery := "INSERT INTO users(id,name) VALUES($1,$2)"
	wantArgs := []interface{}{1, "John"}

	gotQuery, gotArgs, err := GetInsertQuery(&user)

	if err != nil {
		t.Errorf("GetInsertQuery() error = %v, wantErr %v", err, false)
	}

	if gotQuery != wantQuery {
		t.Errorf("GetInsertQuery() query = %v, want %v", gotQuery, wantQuery)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("GetInsertQuery() args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestGetInsertQuery_ErrorIfModelIsNotPointerToStruct(t *testing.T) {
	user := UserTest{Id: 1, Name: "John"}

	_, _, err := GetInsertQuery(user)

	if err != ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT {
		t.Errorf("GetInsertQuery() error = %v, wantErr %v", err, ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT)
	}
}

func TestGetInsertQuery_ErrorIfModelHasNoFields(t *testing.T) {
	user := UserInvalid{}

	_, _, err := GetInsertQuery(&user)

	if err != ERROR_MODEL_HAVE_NO_FIELD {
		t.Errorf("GetInsertQuery() error = %v, wantErr %v", err, ERROR_MODEL_HAVE_NO_FIELD)
	}
}

func TestGetUpdateQuery_Success(t *testing.T) {
	user := &UserTest{Id: 1, Name: "John"}

	wantQuery := "UPDATE users SET name = $1 WHERE id = $2"
	wantArgs := []interface{}{"John", 1}

	gotQuery, gotArgs, err := GetUpdateQuery(user)

	if err != nil {
		t.Errorf("GetUpdateQuery() error = %v, wantErr %v", err, false)
	}

	if gotQuery != wantQuery {
		t.Errorf("GetUpdateQuery() query = %v, want %v", gotQuery, wantQuery)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("GetUpdateQuery() args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestGetUpdateQuery_ErrorIfModelIsNotPointerToStruct(t *testing.T) {
	user := UserTest{Id: 1, Name: "John"}

	_, _, err := GetUpdateQuery(user)

	if err != ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT {
		t.Errorf("GetUpdateQuery() error = %v, wantErr %v", err, ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT)
	}
}

func TestGetUpdateQuery_ErrorIfModelHasNoFields(t *testing.T) {
	user := UserInvalid{}

	_, _, err := GetUpdateQuery(&user)

	if err != ERROR_MODEL_HAVE_NO_FIELD {
		t.Errorf("GetUpdateQuery() error = %v, wantErr %v", err, ERROR_MODEL_HAVE_NO_FIELD)
	}
}

func TestGetUpdateQuery_ErrorIfPrimaryKeyNotFound(t *testing.T) {
	user := &UserTestNotFoundPrimaryKey{
		Id:   1,
		Name: "John",
	}
	_, _, err := GetUpdateQuery(user)

	if err != ERROR_NOT_FOUND_PRIMARY_KEY {
		t.Errorf("GetUpdateQuery() error = %v, wantErr %v", err, ERROR_NOT_FOUND_PRIMARY_KEY)
	}
}

func TestGetDeleteQuery_Success(t *testing.T) {
	user := &UserTest{Id: 1, Name: "John"}

	wantQuery := "DELETE FROM users WHERE id = $1"
	wantArgs := []interface{}{1}

	gotQuery, gotArgs, err := GetDeleteQuery(user)

	if err != nil {
		t.Errorf("GetDeleteQuery() error = %v, wantErr %v", err, false)
	}

	if gotQuery != wantQuery {
		t.Errorf("GetDeleteQuery() query = %v, want %v", gotQuery, wantQuery)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("GetDeleteQuery() args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestGetDeleteQuery_ErrorIfModelIsNotPointerToStruct(t *testing.T) {
	user := UserTest{Id: 1, Name: "John"}

	_, _, err := GetDeleteQuery(user)

	if err == nil {
		t.Errorf("GetDeleteQuery() error = %v, wantErr %v", err, true)
	}
}

func TestGetDeleteQuery_ErrorIfPrimaryKeyNotFound(t *testing.T) {
	user := &UserTestNotFoundPrimaryKey{}
	_, _, err := GetDeleteQuery(user)

	if err != ERROR_NOT_FOUND_PRIMARY_KEY {
		t.Errorf("GetDeleteQuery() error = %v, wantErr %v", err, ERROR_NOT_FOUND_PRIMARY_KEY)
	}
}

func TestGetInsertQueryWithoutPrimaryKey_Success(t *testing.T) {
	user := &UserTest{Id: 1, Name: "John"}

	wantQuery := "INSERT INTO users(name) VALUES($1) RETURNING id"
	wantArgs := []interface{}{"John"}

	gotQuery, gotArgs, _, err := GetInsertQueryWithoutPrimaryKey(user)

	if err != nil {
		t.Errorf("GetInsertQueryWithoutPrimaryKey() error = %v, wantErr %v", err, false)
	}

	if gotQuery != wantQuery {
		t.Errorf("GetInsertQueryWithoutPrimaryKey() query = %v, want %v", gotQuery, wantQuery)
	}

	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Errorf("GetInsertQueryWithoutPrimaryKey() args = %v, want %v", gotArgs, wantArgs)
	}
}

func TestGetInsertQueryWithoutPrimaryKey_ErrorIfModelIsNotPointerToStruct(t *testing.T) {
	user := UserTest{Id: 1, Name: "John"}

	_, _, _, err := GetInsertQueryWithoutPrimaryKey(user)

	if err != ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT {
		t.Errorf("GetInsertQueryWithoutPrimaryKey() error = %v, wantErr %v", err, ERROR_PARAM_IS_NOT_A_POINTER_OF_STRUCT)
	}
}

func TestGetInsertQueryWithoutPrimaryKey_ErrorIfModelHasNoFields(t *testing.T) {
	user := UserInvalid{}

	_, _, _, err := GetInsertQueryWithoutPrimaryKey(&user)

	if err != ERROR_MODEL_HAVE_NO_FIELD {
		t.Errorf("GetInsertQueryWithoutPrimaryKey() error = %v, wantErr %v", err, ERROR_MODEL_HAVE_NO_FIELD)
	}
}

func TestGetInsertQueryWithoutPrimaryKey_ErrorIfPrimaryKeyNotFound(t *testing.T) {
	user := &UserTestNotFoundPrimaryKey{
		Id:   1,
		Name: "John",
	}
	_, _, _, err := GetInsertQueryWithoutPrimaryKey(user)

	if err != ERROR_NOT_FOUND_PRIMARY_KEY {
		t.Errorf("GetInsertQueryWithoutPrimaryKey() error = %v, wantErr %v", err, ERROR_NOT_FOUND_PRIMARY_KEY)
	}
}
