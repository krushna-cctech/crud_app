package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func mockDatabase(t *testing.T) sqlmock.Sqlmock {
	t.Helper()

	previousDB := db
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	db = mockDB
	t.Cleanup(func() {
		db = previousDB
		mockDB.Close()
	})

	return mock
}

func TestCreateUser(t *testing.T) {
	mock := mockDatabase(t)
	expect := mock.ExpectExec(regexp.QuoteMeta("INSERT INTO users (name, email) VALUES (?, ?)"))
	expect.WithArgs("Ada Lovelace", "ada@example.com").WillReturnResult(sqlmock.NewResult(7, 1))

	request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Ada Lovelace","email":"ada@example.com"}`))
	response := httptest.NewRecorder()
	usersHandler(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body)
	}
	var user User
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.ID != 7 || user.Name != "Ada Lovelace" || user.Email != "ada@example.com" {
		t.Errorf("response user = %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations: %v", err)
	}
}

func TestGetUsers(t *testing.T) {
	mock := mockDatabase(t)
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(7, "Ada Lovelace", "ada@example.com", "2025-01-02 03:04:05")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, created_at FROM users")).WillReturnRows(rows)

	request := httptest.NewRequest(http.MethodGet, "/users", nil)
	response := httptest.NewRecorder()
	usersHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body)
	}
	var users []User
	if err := json.NewDecoder(response.Body).Decode(&users); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(users) != 1 || users[0].ID != 7 {
		t.Errorf("response users = %+v, want one user with ID 7", users)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations: %v", err)
	}
}

func TestGetUser(t *testing.T) {
	mock := mockDatabase(t)
	rows := sqlmock.NewRows([]string{"id", "name", "email", "created_at"}).
		AddRow(7, "Ada Lovelace", "ada@example.com", "2025-01-02 03:04:05")
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, email, created_at FROM users WHERE id = ?")).
		WithArgs(7).WillReturnRows(rows)

	request := httptest.NewRequest(http.MethodGet, "/users/7", nil)
	response := httptest.NewRecorder()
	userHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body)
	}
	var user User
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.ID != 7 {
		t.Errorf("response user ID = %d, want 7", user.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	mock := mockDatabase(t)
	expect := mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET name = ?, email = ? WHERE id = ?"))
	expect.WithArgs("Ada Byron", "ada@example.com", 7).WillReturnResult(sqlmock.NewResult(0, 1))

	request := httptest.NewRequest(http.MethodPut, "/users/7", strings.NewReader(`{"name":"Ada Byron","email":"ada@example.com"}`))
	response := httptest.NewRecorder()
	userHandler(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body)
	}
	var user User
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if user.ID != 7 || user.Name != "Ada Byron" {
		t.Errorf("response user = %+v", user)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations: %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	mock := mockDatabase(t)
	expect := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE id = ?"))
	expect.WithArgs(7).WillReturnResult(sqlmock.NewResult(0, 1))

	request := httptest.NewRequest(http.MethodDelete, "/users/7", nil)
	response := httptest.NewRecorder()
	userHandler(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusNoContent, response.Body)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("SQL expectations: %v", err)
	}
}

func TestUserHandlerRejectsInvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/users/not-a-number", nil)
	response := httptest.NewRecorder()
	userHandler(response, request)

	if response.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
