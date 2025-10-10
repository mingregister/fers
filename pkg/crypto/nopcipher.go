package crypto

type nopCipher struct{}

var _ Cipher = (*nopCipher)(nil)

func NewNopCipher() Cipher {
	return &nopCipher{}
}

func (nc *nopCipher) Encrypt(plain []byte) ([]byte, error) {
	// No encryption, return the original data
	return plain, nil
}
func (nc *nopCipher) Decrypt(cipher []byte) ([]byte, error) {
	// No decryption, return the original data
	return cipher, nil
}
