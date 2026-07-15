package auth

type sessionRow struct {
	userID    string
	expiresAt int64
}

func (s *Service) createSession(token, userID string, now, expires int64) error {
	_, err := s.db.Exec(`INSERT INTO sessions (token, user_id, created_at, expires_at) VALUES (?, ?, ?, ?)`, token, userID, now, expires)
	return err
}

func (s *Service) lookupSession(token string) (sessionRow, bool) {
	var r sessionRow
	err := s.db.QueryRow(
		`SELECT user_id, expires_at FROM sessions WHERE token = ?`,
		token,
	).Scan(&r.userID, &r.expiresAt)
	if err != nil {
		return sessionRow{}, false
	}
	return r, true
}

func (s *Service) deleteSession(token string) {
	s.db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
}

func (s *Service) deleteExpiredSessions(now int64) {
	s.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, now)
}
