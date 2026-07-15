package agentkey

import (
	"fmt"
	"time"

	"agentmanager/modules/crypto"
)

// Validate resolves an agent token to the account that owns it. From the
// agent's side nothing changes — it still presents a bearer token — but the
// server now learns which account that token belongs to.
func (s *Service) Validate(token string) (userID string, ok bool) {
	if token == "" {
		return "", false
	}
	return s.getByHash(crypto.SHA256Hex(token))
}

func (s *Service) Info(userID string) (exists bool, createdAt int64) {
	row, ok := s.getByUser(userID)
	return ok, row.createdAt
}

// Rotate mints a fresh key for one account, replacing any existing one. It's
// used both to issue the initial key at registration and to rotate later.
// onRotate fires with the account id so the hub can drop that account's live
// agent connections, forcing them to reconnect with the new key.
func (s *Service) Rotate(userID string) (raw string, err error) {
	raw = crypto.RandomHex(32)
	if err = s.set(userID, crypto.SHA256Hex(raw), raw, time.Now().Unix()); err != nil {
		return "", err
	}
	if s.onRotate != nil {
		s.onRotate(userID)
	}
	return raw, nil
}

func (s *Service) RawKey(userID string) (string, error) {
	raw, ok := s.getRaw(userID)
	if !ok || raw == "" {
		return "", fmt.Errorf("no key found")
	}
	return raw, nil
}
