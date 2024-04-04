package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMongoGetDB(t *testing.T) {
	// Test get mongo db
	info := mongoDBInfo{
		Host:     "localhost",
		Port:     27017,
		Username: "",
		Password: "",
		Database: "coin_test",
	}

	session := openMongoDBConnection(info)

	assert.NotNil(t, session, "MongoDB database should not be nil")

	t.Logf("Get successfully mongo db: %#v", info)
}
