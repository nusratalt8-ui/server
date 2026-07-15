//go:build windows

package loader

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/microsoft/UpdateAssistant/modules/config"
)

func decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(config.PluginKey))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := gcm.NonceSize()
	if len(data) < ns {
		return nil, ErrLoadFailed
	}
	return gcm.Open(nil, data[:ns], data[ns:], nil)
}
