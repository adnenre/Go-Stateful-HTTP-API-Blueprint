package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"

	"testing"
	"time"

	"rest-api-blueprint/internal/errors"
)

func TestContainerMiddlewareChain(t *testing.T) {
	baseURL := "http://localhost:8080/api/v1/health"

	// Check if the container is reachable
	resp, err := http.Get(baseURL)
	if err != nil {
		t.Skipf("Container not reachable at %s. Run 'make docker-dev' first.", baseURL)
	}
	resp.Body.Close()

	// Helper to flush Redis (so rate limit tests are deterministic)
	flushRedis := func() {
		cmd := exec.Command("docker", "exec", "rest_api_redis", "redis-cli", "FLUSHALL")
		cmd.Run() // ignore error; if Redis not accessible, test may still pass
	}

	// Helper to make a request and return response, body, and request ID
	makeRequest := func(customID string) (*http.Response, []byte, string) {
		req, _ := http.NewRequest("GET", baseURL, nil)
		if customID != "" {
			req.Header.Set("X-Request-ID", customID)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)
		headerID := resp.Header.Get("X-Request-ID")
		return resp, body.Bytes(), headerID
	}

	// ============================================================
	// SCENARIO 1: No client header → generated request ID
	// ============================================================
	t.Run("generated_request_id", func(t *testing.T) {
		flushRedis()
		// Wait for rate limit window to reset (if needed)
		time.Sleep(1 * time.Second)

		// Make two successful requests (should be 200)
		for i := 0; i < 2; i++ {
			resp, _, id := makeRequest("")
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}
			if id == "" {
				t.Fatal("X-Request-ID header missing")
			}
		}
		// Third request should be 429
		resp, body, headerID := makeRequest("")
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", resp.StatusCode)
		}
		if headerID == "" {
			t.Fatal("X-Request-ID header missing for 429")
		}
		var problem errors.ProblemDetails
		if err := json.Unmarshal(body, &problem); err != nil {
			t.Fatal(err)
		}
		if problem.Instance != headerID {
			t.Errorf("instance = %s, wanted %s", problem.Instance, headerID)
		}
	})

	// ============================================================
	// SCENARIO 2: Client provides X-Request-ID header
	// ============================================================
	t.Run("client_provided_request_id", func(t *testing.T) {
		flushRedis()
		time.Sleep(1 * time.Second)

		customID := "my-e2e-test-123"
		// Two successful requests
		for i := 0; i < 2; i++ {
			resp, _, id := makeRequest(customID)
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("expected 200, got %d", resp.StatusCode)
			}
			if id != customID {
				t.Errorf("header = %s, want %s", id, customID)
			}
		}
		// Third request – 429
		resp, body, headerID := makeRequest(customID)
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", resp.StatusCode)
		}
		if headerID != customID {
			t.Errorf("header = %s, want %s", headerID, customID)
		}
		var problem errors.ProblemDetails
		if err := json.Unmarshal(body, &problem); err != nil {
			t.Fatal(err)
		}
		if problem.Instance != customID {
			t.Errorf("instance = %s, want %s", problem.Instance, customID)
		}
	})
}
