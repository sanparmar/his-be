package domain

import "testing"

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}

	if hash == password {
		t.Fatal("HashPassword returned plain password")
	}

	// Verify the hash
	if err := VerifyPassword(hash, password); err != nil {
		t.Fatalf("VerifyPassword failed for correct password: %v", err)
	}

	// Wrong password should fail
	if err := VerifyPassword(hash, "wrongpassword"); err == nil {
		t.Fatal("VerifyPassword succeeded for wrong password")
	}
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	err := VerifyPassword("$2a$12$hash", "")
	if err == nil {
		t.Fatal("VerifyPassword should fail for empty password")
	}
}
