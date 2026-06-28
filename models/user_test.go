package models

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	user := &User{}
	err := user.HashPassword("mypassword123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if user.Password == "" {
		t.Fatal("Password should not be empty after hashing")
	}

	if user.Password == "mypassword123" {
		t.Fatal("Password should be hashed, not plaintext")
	}

	if !strings.HasPrefix(user.Password, "$2a$") {
		t.Fatalf("Password should start with bcrypt prefix $2a$, got: %s", user.Password[:6])
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	user1 := &User{}
	user2 := &User{}

	err := user1.HashPassword("samepassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	err = user2.HashPassword("samepassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if user1.Password == user2.Password {
		t.Fatal("Same password should produce different hashes (random salt)")
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	user := &User{}
	err := user.HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if !user.CheckPassword("correctpassword") {
		t.Fatal("CheckPassword should return true for correct password")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	user := &User{}
	err := user.HashPassword("correctpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if user.CheckPassword("wrongpassword") {
		t.Fatal("CheckPassword should return false for wrong password")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	user := &User{}
	err := user.HashPassword("password123")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if user.CheckPassword("") {
		t.Fatal("CheckPassword should return false for empty password")
	}
}
