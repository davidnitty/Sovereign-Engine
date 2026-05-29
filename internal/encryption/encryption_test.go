package encryption

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	svc, err := New("test-master-key")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	ciphertext, err := svc.Encrypt("secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if ciphertext == "secret" {
		t.Fatal("ciphertext must not equal plaintext")
	}
	plaintext, err := svc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "secret" {
		t.Fatalf("plaintext = %q, want secret", plaintext)
	}
}
