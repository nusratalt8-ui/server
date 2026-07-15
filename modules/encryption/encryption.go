package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"sync"
)

var ErrDecryptionFailed = errors.New("decryption failed")

const chunkSize = 64 * 1024

var keyCache sync.Map

func deriveKey(masterKey string) []byte {
	if v, ok := keyCache.Load(masterKey); ok {
		return v.([]byte)
	}
	h := sha256.Sum256([]byte(masterKey))
	keyCache.Store(masterKey, h[:])
	return h[:]
}

func encryptWithKey(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decryptWithKey(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return nil, ErrDecryptionFailed
	}
	plain, err := gcm.Open(nil, data[:ns], data[ns:], nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	return plain, nil
}

func EncryptBytes(data []byte, masterKey string) ([]byte, error) {
	key := deriveKey(masterKey)
	if len(data) <= chunkSize {
		return encryptWithKey(data, key)
	}
	return encryptChunked(data, key)
}

func DecryptBytes(data []byte, masterKey string) ([]byte, error) {
	key := deriveKey(masterKey)
	if len(data) > 4 && data[0] == 'C' && data[1] == 'H' && data[2] == 'K' && data[3] == 1 {
		return decryptChunked(data, key)
	}
	return decryptWithKey(data, key)
}

func encryptChunked(data, key []byte) ([]byte, error) {
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	ns := gcm.NonceSize()
	out := []byte{'C', 'H', 'K', 1}
	for off := 0; off < len(data); off += chunkSize {
		end := off + chunkSize
		if end > len(data) {
			end = len(data)
		}
		nonce := make([]byte, ns)
		rand.Read(nonce)
		out = append(out, gcm.Seal(nonce, nonce, data[off:end], nil)...)
	}
	return out, nil
}

func decryptChunked(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	gcm, _ := cipher.NewGCM(block)
	ns := gcm.NonceSize()
	overhead := gcm.Overhead()
	chunkEnc := ns + chunkSize + overhead
	payload := data[4:]
	var out []byte
	for len(payload) > 0 {
		sz := chunkEnc
		if sz > len(payload) {
			sz = len(payload)
		}
		chunk := payload[:sz]
		payload = payload[sz:]
		if len(chunk) < ns {
			return nil, ErrDecryptionFailed
		}
		plain, err := gcm.Open(nil, chunk[:ns], chunk[ns:], nil)
		if err != nil {
			return nil, ErrDecryptionFailed
		}
		out = append(out, plain...)
	}
	return out, nil
}
