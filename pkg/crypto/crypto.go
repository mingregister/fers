package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
)

// 加密行为
type Encrypter interface {
	Encrypt(plain []byte) (cipher []byte, err error)
}

// 解密行为
type Decrypter interface {
	Decrypt(cipher []byte) (plain []byte, err error)
}

// 加解密一体
type Cipher interface {
	Encrypter
	Decrypter
}

type aesGCM struct {
	key []byte
}

func NewAESGCM(password string) Cipher {
	h := sha256.Sum256([]byte(password))
	return &aesGCM{key: h[:]}
}

var _ Cipher = &aesGCM{}

func (ag *aesGCM) Encrypt(plain []byte) ([]byte, error) {
	block, err := aes.NewCipher(ag.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nil, nonce, plain, nil)
	// store nonce + ciphertext
	out := append(nonce, ciphertext...)
	return out, nil
}

func (ag *aesGCM) Decrypt(cipherData []byte) (plain []byte, err error) {
	block, err := aes.NewCipher(ag.key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(cipherData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := cipherData[:nonceSize]
	ct := cipherData[nonceSize:]
	return gcm.Open(nil, nonce, ct, nil)
}
