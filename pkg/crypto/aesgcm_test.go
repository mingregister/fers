package crypto

import (
	"bytes"
	"testing"
)

func TestNewAESGCM(t *testing.T) {
	password := "test-password"
	cipher := NewAESGCM(password)

	if cipher == nil {
		t.Fatal("NewAESGCM returned nil")
	}

	// Test that the same password creates consistent cipher
	cipher2 := NewAESGCM(password)
	if cipher2 == nil {
		t.Fatal("NewAESGCM returned nil for second instance")
	}
}

func TestAESGCM_EncryptDecrypt(t *testing.T) {
	password := "test-password-123"
	cipher := NewAESGCM(password)

	testCases := []struct {
		name      string
		plaintext []byte
	}{
		{
			name:      "empty data",
			plaintext: []byte{},
		},
		{
			name:      "simple text",
			plaintext: []byte("hello world"),
		},
		{
			name:      "binary data",
			plaintext: []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
		},
		{
			name:      "large text",
			plaintext: bytes.Repeat([]byte("test data "), 1000),
		},
		{
			name:      "unicode text",
			plaintext: []byte("测试数据 🔐 encryption"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := cipher.Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Verify encrypted data is different from original
			if len(tc.plaintext) > 0 && bytes.Equal(encrypted, tc.plaintext) {
				t.Error("Encrypted data should be different from plaintext")
			}

			// Decrypt
			decrypted, err := cipher.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			// Verify decrypted data matches original
			if !bytes.Equal(decrypted, tc.plaintext) {
				t.Errorf("Decrypted data doesn't match original.\nExpected: %v\nGot: %v", tc.plaintext, decrypted)
			}
		})
	}
}

func TestAESGCM_EncryptionUniqueness(t *testing.T) {
	password := "test-password"
	cipher := NewAESGCM(password)
	plaintext := []byte("same data")

	// Encrypt the same data multiple times
	encrypted1, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	encrypted2, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// Encrypted data should be different each time due to random nonce
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Encrypted data should be different each time due to random nonce")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := cipher.Decrypt(encrypted1)
	if err != nil {
		t.Fatalf("First decryption failed: %v", err)
	}

	decrypted2, err := cipher.Decrypt(encrypted2)
	if err != nil {
		t.Fatalf("Second decryption failed: %v", err)
	}

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Both decryptions should produce the original plaintext")
	}
}

func TestAESGCM_DecryptInvalidData(t *testing.T) {
	password := "test-password"
	cipher := NewAESGCM(password)

	testCases := []struct {
		name        string
		invalidData []byte
	}{
		{
			name:        "empty data",
			invalidData: []byte{},
		},
		{
			name:        "too short data",
			invalidData: []byte{0x01, 0x02},
		},
		{
			name:        "random data",
			invalidData: []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F, 0x10},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := cipher.Decrypt(tc.invalidData)
			if err == nil {
				t.Error("Decrypt should fail with invalid data")
			}
		})
	}
}

func TestAESGCM_DifferentPasswords(t *testing.T) {
	plaintext := []byte("secret message")

	cipher1 := NewAESGCM("password1")
	cipher2 := NewAESGCM("password2")

	// Encrypt with first cipher
	encrypted, err := cipher1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	// Try to decrypt with second cipher (different password)
	_, err = cipher2.Decrypt(encrypted)
	if err == nil {
		t.Error("Decryption should fail when using different password")
	}

	// Verify first cipher can still decrypt correctly
	decrypted, err := cipher1.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption with correct password failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted data doesn't match original")
	}
}

func TestAESGCM_InterfaceCompliance(t *testing.T) {
	cipher := NewAESGCM("test")

	// Test that it implements all required interfaces
	var _ Cipher = cipher
	var _ Encrypter = cipher
	var _ Decrypter = cipher
}

func BenchmarkAESGCM_Encrypt(b *testing.B) {
	cipher := NewAESGCM("benchmark-password")
	data := bytes.Repeat([]byte("benchmark data "), 100) // ~1.5KB

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cipher.Encrypt(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAESGCM_Decrypt(b *testing.B) {
	cipher := NewAESGCM("benchmark-password")
	data := bytes.Repeat([]byte("benchmark data "), 100) // ~1.5KB

	encrypted, err := cipher.Encrypt(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cipher.Decrypt(encrypted)
		if err != nil {
			b.Fatal(err)
		}
	}
}
