package api_test

import (
	"net/http"
	"strings"
	"testing"

	"cosign/internal/api"
	"cosign/internal/service"
	"cosign/internal/util"
)

func TestCreateAPIKeyEndpoint(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	// post create key
	var response struct {
		Error api.APIError `json:"error"`
		Token string       `json:"data"`
	}
	auth := makeTestAuthHeader(t)
	result := post(router, "/api/v1/settings/keys", "", &response, auth)

	// validate result
	expectStatus(t, http.StatusCreated, result)

	// validate response
	if response.Token == "" {
		t.Fatalf("expected token in response")
	}

	// validate resource creation
	ok, err := service.VerifyAPIKey(response.Token)
	if err != nil || !ok {
		t.Fatalf("token verification failed: %v", err)
	}
}

func TestDeleteAPIKeyEndpoint(t *testing.T) {

	util.SetupTestDB(t)
	router := setupRouter()

	// create API key
	token, err := service.CreateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	id := strings.Split(token, ".")[0]

	// del key id
	auth := makeTestAuthHeader(t)
	result := del(router, "/api/v1/settings/keys/"+id, nil, auth)

	// validate result
	expectStatus(t, http.StatusNoContent, result)

	// validate resource deletion
	_, err = service.VerifyAPIKey(token)
	if err == nil {
		t.Fatalf("VerifyAPIKey should fail after deletion")
	}
}
