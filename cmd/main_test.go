package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sirius_future/internal"
	"sirius_future/internal/app/entity"
	"sirius_future/internal/app/repository"
	"sirius_future/internal/app/service"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const adminToken = "test-admin-token"

func setup(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	db, err := internal.DatabaseInit(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	return newApp(db, config{AdminToken: adminToken}), db
}

func call(t *testing.T, app *fiber.App, method, path string, admin bool, body any) (int, []byte) {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	if admin {
		req.Header.Set("Authorization", "Bearer "+adminToken)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

func user(n int, url string) map[string]string {
	return map[string]string{
		"first_name":  "User",
		"second_name": "Test",
		"last_name":   "Name",
		"email":       fmt.Sprintf("user%d@example.com", n),
		"password":    "password123",
		"phone":       "+998901234567",
		"url":         url,
	}
}

func register(t *testing.T, app *fiber.App, n int, url string) (int, map[string]any) {
	t.Helper()
	status, raw := call(t, app, http.MethodPost, "/register", false, user(n, url))
	out := map[string]any{}
	json.Unmarshal(raw, &out)
	return status, out
}

func createLink(t *testing.T, app *fiber.App, userID uint, limit int) string {
	t.Helper()
	status, raw := call(t, app, http.MethodPost, "/api/create-link", true, map[string]any{"user_id": userID, "link_limit": limit})
	if status != http.StatusCreated {
		t.Fatalf("create link: %d %s", status, raw)
	}
	var out struct {
		Link string `json:"link"`
	}
	json.Unmarshal(raw, &out)
	return out.Link
}

func TestAdminRoutesNeedToken(t *testing.T) {
	app, _ := setup(t)
	for _, path := range []string{"/api/users", "/api/links", "/api/payments"} {
		if status, _ := call(t, app, http.MethodGet, path, false, nil); status != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", path, status)
		}
	}
}

func TestPasswordIsHashedAndHidden(t *testing.T) {
	app, db := setup(t)
	if status, body := register(t, app, 1, ""); status != http.StatusCreated {
		t.Fatalf("register: %d %v", status, body)
	}

	var u entity.User
	db.First(&u)
	if u.Password == "password123" || u.Password == "" {
		t.Fatal("password is not hashed")
	}
	if u.Role != "user" {
		t.Fatalf("role = %q", u.Role)
	}

	_, raw := call(t, app, http.MethodGet, "/api/users", true, nil)
	if bytes.Contains(raw, []byte(u.Password)) || bytes.Contains(raw, []byte("password")) {
		t.Fatalf("password leaked: %s", raw)
	}
}

func TestReferralLimit(t *testing.T) {
	app, _ := setup(t)
	_, owner := register(t, app, 1, "")
	ownerID := uint(owner["id"].(float64))
	link := createLink(t, app, ownerID, 2)

	for i := 2; i <= 3; i++ {
		status, body := register(t, app, i, link)
		if status != http.StatusCreated {
			t.Fatalf("referral %d: %d %v", i, status, body)
		}
		if uint(body["referrer_id"].(float64)) != ownerID {
			t.Fatalf("wrong referrer: %v", body)
		}
	}

	if status, _ := register(t, app, 4, link); status != http.StatusConflict {
		t.Fatalf("expected 409 after limit, got %d", status)
	}
	if status, raw := call(t, app, http.MethodGet, "/api/check-link/"+link, false, nil); status != http.StatusOK || !bytes.Contains(raw, []byte("false")) {
		t.Fatalf("check exhausted link: %d %s", status, raw)
	}
	if status, _ := register(t, app, 5, "missing-link"); status != http.StatusNotFound {
		t.Fatalf("unknown link: expected 404, got %d", status)
	}
}

func TestReferralLimitUnderConcurrency(t *testing.T) {
	app, db := setup(t)
	_, owner := register(t, app, 1, "")
	link := createLink(t, app, uint(owner["id"].(float64)), 5)

	repo := repository.NewFutureSiriusRepository(db, service.NewLoggerService(service.InitLogger()))

	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u := &entity.User{
				Firstname:  "A",
				Secondname: "B",
				Lastname:   "C",
				Email:      fmt.Sprintf("c%d@example.com", i),
				Password:   "hash",
				Phone:      "+998901234567",
				Role:       "user",
			}
			if err := repo.CreateUserWithReferral(u, link); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	if ok != 5 {
		t.Fatalf("expected exactly 5 registrations by link with limit 5, got %d", ok)
	}
}

func TestPayments(t *testing.T) {
	app, _ := setup(t)
	_, u := register(t, app, 1, "")
	id := uint(u["id"].(float64))

	if status, raw := call(t, app, http.MethodPost, "/api/payments", true, map[string]any{"user_id": id, "amount": 100, "description": "lesson"}); status != http.StatusCreated {
		t.Fatalf("create payment: %d %s", status, raw)
	}
	if status, _ := call(t, app, http.MethodPost, "/api/payments", true, map[string]any{"user_id": 999, "amount": 100}); status != http.StatusBadRequest {
		t.Fatalf("payment for unknown user: expected 400, got %d", status)
	}

	_, raw := call(t, app, http.MethodGet, fmt.Sprintf("/api/payments/user/%d", id), true, nil)
	var payments []entity.Payment
	json.Unmarshal(raw, &payments)
	if len(payments) != 1 || payments[0].Status != "pending" {
		t.Fatalf("unexpected payments: %s", raw)
	}
}

func TestMetrics(t *testing.T) {
	app, _ := setup(t)
	call(t, app, http.MethodGet, "/api/check-link/abc", false, nil)

	status, raw := call(t, app, http.MethodGet, "/metrics", false, nil)
	if status != http.StatusOK || !bytes.Contains(raw, []byte(`path="/api/check-link/:url"`)) {
		t.Fatalf("metrics: %d", status)
	}
}
