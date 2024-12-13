package core

import (
	"net/http"
	"testing"
)

func TestInternalHttpClientRequest_NormalRequest(t *testing.T) {
	client := NewClient()

	ctx := GetContext()

	type responseStruct struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	responseValue := &responseStruct{}

	res, err := client.
		SetUrl("http://localhost:1080/core/http-internal-test").
		AddHeader("Content-Type", "application/json").
		SetMethod(http.MethodPost).
		SetBody(nil).
		SetContext(ctx).
		RequestInternal(responseValue)

	if err != nil {
		t.Errorf("Error: %v", err)
		return
	}

	if res.GetStatusCode() != http.StatusOK {
		t.Errorf("Status code is not 200: %v", res.GetStatusCode())
		return
	}

	t.Logf("Response: %v", responseValue)
}
