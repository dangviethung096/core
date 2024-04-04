package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenMongoDBConnection(t *testing.T) {
	// Test open mongo db connection
	info := mongoDBInfo{
		Host:     "localhost",
		Port:     27017,
		Username: "",
		Password: "",
		Database: "coin_test",
	}

	session := openMongoDBConnection(info)

	assert.NotNil(t, session.client, "MongoDB client should not be nil")

	t.Logf("Connect successfully to mongo db: %#v", info)
}
