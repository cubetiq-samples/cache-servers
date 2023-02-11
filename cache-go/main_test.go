package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestRoutes(t *testing.T) {
	// Start a local HTTP server
	router := NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	// Test POST /cache?key={key}&value={value}
	resp, err := http.Post(server.URL+"/cache?key=test&value=test_value", "", nil)
	if err != nil {
		t.Fatalf("Error sending POST request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	// Test GET /cache/keys
	resp, err = http.Get(server.URL + "/cache/keys")
	if err != nil {
		t.Fatalf("Error sending GET request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	// Test GET /cache?key={key}
	resp, err = http.Get(server.URL + "/cache?key=test")
	if err != nil {
		t.Fatalf("Error sending GET request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	// Test DELETE /cache?key={key}
	req, err := http.NewRequest(http.MethodDelete, server.URL+"/cache?key=test", nil)
	if err != nil {
		t.Fatalf("Error creating DELETE request: %v", err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error sending DELETE request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func BenchmarkTest(b *testing.B) {
	// Start a local HTTP server
	router := NewRouter()
	server := httptest.NewServer(router)
	defer server.Close()

	for i := 0; i < b.N; i++ {
		// Test POST /cache?key={key}&value={value}
		key := "test" + strconv.Itoa(i)
		value := "test_value" + strconv.Itoa(i)

		resp, err := http.Post(server.URL+"/cache?key="+key+"&value="+value, "", nil)
		if err != nil {
			b.Fatalf("Error sending POST request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			b.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
		}

		// Test GET /cache?key={key}
		resp, err = http.Get(server.URL + "/cache?key=" + key)
		if err != nil {
			b.Fatalf("Error sending GET request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			b.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
		}

		// Test DELETE /cache?key={key}
		req, err := http.NewRequest(http.MethodDelete, server.URL+"/cache?key="+key, nil)
		if err != nil {

			b.Fatalf("Error creating DELETE request: %v", err)
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("Error sending DELETE request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			b.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
		}
	}
}
