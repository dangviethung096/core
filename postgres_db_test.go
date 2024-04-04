package core

import "testing"

func TestOpenDBWithSuccessResponse(t *testing.T) {
	// Set up test cases
	testCase := struct {
		name    string
		dbInfo  postgresDBInfo
		wantErr bool
	}{
		name: "Valid DBInfo",
		dbInfo: postgresDBInfo{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "hungdeptrai",
			Database: "account",
		},
		wantErr: false,
	}

	// Call the function being tested
	db := openPostgresDBConnection(testCase.dbInfo)
	db.Close()
}
