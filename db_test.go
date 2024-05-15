package core

import "testing"

func TestOpenDBWithSuccessResponse(t *testing.T) {
	// Set up test cases
	testCase := struct {
		name    string
		dbInfo  DBInfo
		wantErr bool
	}{
		name: "Valid DBInfo",
		dbInfo: DBInfo{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "admin@123",
			Database: "example",
		},
		wantErr: false,
	}

	// Call the function being tested
	db := openDBConnection(testCase.dbInfo)
	db.Close()
}
