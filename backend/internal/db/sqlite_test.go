package db

import (
	"path/filepath"
	"testing"
)

func TestNewDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()
}

func TestCreateAndGetUser(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	user, err := d.CreateUser("testuser", "hash123", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("username = %q, want %q", user.Username, "testuser")
	}
	if user.ReadOnly != false {
		t.Errorf("readOnly = %v, want false", user.ReadOnly)
	}
	if user.ID == 0 {
		t.Errorf("ID should not be zero")
	}

	got, err := d.GetUserByUsername("testuser")
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}
	if got.Username != user.Username || got.ID != user.ID {
		t.Errorf("got %+v, want %+v", got, user)
	}
}

func TestCreateAndGetSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	user, err := d.CreateUser("sessionuser", "hash", false)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	err = d.CreateSession(user.ID, "token-123", "2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}

	s, err := d.GetSessionByToken("token-123")
	if err != nil {
		t.Fatalf("GetSessionByToken() error = %v", err)
	}
	if s.UserID != user.ID || s.Token != "token-123" {
		t.Errorf("got %+v", s)
	}
}

func TestDeleteSession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	user, _ := d.CreateUser("deleteuser", "hash", false)
	d.CreateSession(user.ID, "token-delete", "2025-01-01T00:00:00Z")

	err = d.DeleteSession("token-delete")
	if err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}

	_, err = d.GetSessionByToken("token-delete")
	if err == nil {
		t.Error("expected error for deleted session")
	}
}

func TestEnsureAdminCreatesAdmin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	err = d.EnsureAdmin("admin", "password")
	if err != nil {
		t.Fatalf("EnsureAdmin() error = %v", err)
	}

	user, err := d.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}
	if user.Username != "admin" {
		t.Errorf("username = %q", user.Username)
	}
}

func TestEnsureAdminIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	err = d.EnsureAdmin("admin", "password")
	if err != nil {
		t.Fatalf("first EnsureAdmin() error = %v", err)
	}

	err = d.EnsureAdmin("admin", "otherpass")
	if err != nil {
		t.Fatalf("second EnsureAdmin() error = %v", err)
	}

	_, err = d.GetUserByUsername("admin")
	if err != nil {
		t.Fatalf("GetUserByUsername() error = %v", err)
	}
}

func TestDuplicateUser(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	d.CreateUser("dupuser", "hash1", false)
	_, err = d.CreateUser("dupuser", "hash2", false)
	if err == nil {
		t.Error("expected error for duplicate user")
	}
}

func TestGetNonExistentUser(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := New(path)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer d.Close()

	_, err = d.GetUserByUsername("nonexistent")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
