//go:build windows

package loader

import "io/fs"

func loadBytes(src fs.FS, name string) ([]byte, error) {
	enc, err := fs.ReadFile(src, name+".dll")
	if err != nil {
		return nil, err
	}
	return decrypt(enc)
}
