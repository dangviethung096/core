package core

import (
	"testing"
)

func TestMain(m *testing.M) {
	Init("./example-core.config.yaml")
	m.Run()
}
