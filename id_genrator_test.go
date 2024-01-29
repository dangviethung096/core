package core

import "testing"

func TestGenrateId(t *testing.T) {
	id := ID.GenerateID()
	t.Log(id)
}
