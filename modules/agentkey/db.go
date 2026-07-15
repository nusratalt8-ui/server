package agentkey

type keyRow struct {
	hash      string
	createdAt int64
}

func (s *Service) getByUser(userID string) (keyRow, bool) {
	var r keyRow
	err := s.db.QueryRow(`SELECT key_hash, created_at FROM agent_key WHERE user_id = ?`, userID).
		Scan(&r.hash, &r.createdAt)
	if err != nil {
		return keyRow{}, false
	}
	return r, true
}

func (s *Service) getByHash(hash string) (string, bool) {
	var userID string
	err := s.db.QueryRow(`SELECT user_id FROM agent_key WHERE key_hash = ?`, hash).
		Scan(&userID)
	if err != nil {
		return "", false
	}
	return userID, true
}

func (s *Service) set(userID, hash, raw string, now int64) error {
	_, err := s.db.Exec(
		`INSERT INTO agent_key (user_id, key_hash, key_raw, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET key_hash = excluded.key_hash, key_raw = excluded.key_raw, created_at = excluded.created_at`,
		userID, hash, raw, now,
	)
	return err
}

func (s *Service) getRaw(userID string) (string, bool) {
	var raw string
	err := s.db.QueryRow(`SELECT key_raw FROM agent_key WHERE user_id = ?`, userID).Scan(&raw)
	return raw, err == nil
}
