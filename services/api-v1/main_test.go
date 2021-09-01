package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Gogistics/prj-envoy-v1/services/api-v1/types"
	"github.com/google/go-cmp/cmp"
)

func TestRouter(t *testing.T) {
	// Instantiate the router using the constructor function that
	// we defined previously
	r := newRouter()

	// Create a new server using the "httptest" libraries `NewServer` method
	// Documentation : https://golang.org/pkg/net/http/httptest/#NewServer
	mockServer := httptest.NewServer(r)

	// The mock server we created runs a server and exposes its location in the
	// URL attribute
	// We make a GET request to the "/api/v1" route we defined in the router
	resp, err := http.Get(mockServer.URL + "/api/v1")

	// Handle any unexpected error
	if err != nil {
		t.Fatal(err)
	}

	// We want our status to be 200 (ok)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status should be ok, got %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	// new a Profile struct
	var p types.Profile

	// Try to decode the request body into the struct. If there is an error,
	// throw an error.
	errOfDecode := json.NewDecoder(resp.Body).Decode(&p)
	if errOfDecode != nil {
		t.Errorf("Failed to decode")
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	expected := types.Profile{
		Hostname:      hostname,
		RemoteAddress: "remoteAddr",
		Author:        "Alan",
		Hobbies:       []string{"workout", "programming", "driving"}}
	expectedProfile, err := json.Marshal(expected)
	// We want our response to match the one defined in our handler.
	if cmp.Equal(p, expectedProfile) {
		t.Errorf("Response should be %s, got %s", expectedProfile, p)
	}

}
