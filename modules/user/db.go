package user

import "agentmanager/modules/crypto"

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt int64  `json:"created_at"`
}

func (s *Service) Create(username, passwordHash string, now int64) (string, error) {
	id := crypto.RandomHex(12)
	_, err := s.db.Exec(
		`INSERT INTO users (id, username, password_hash, created_at) VALUES (?, ?, ?, ?)`,
		id, username, passwordHash, now,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) Get(id string) (User, bool) {
	var u User
	err := s.db.QueryRow(`SELECT id, username, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Username, &u.CreatedAt)
	if err != nil {
		return User{}, false
	}
	return u, true
}

func (s *Service) GetCredentials(username string) (id, passwordHash string, ok bool) {
	err := s.db.QueryRow(`SELECT id, password_hash FROM users WHERE username = ?`, username).
		Scan(&id, &passwordHash)
	if err != nil {
		return "", "", false
	}
	return id, passwordHash, true
}

func (s *Service) Any() bool {
	var exists int
	if err := s.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users)`).Scan(&exists); err != nil {
		return false
	}
	return exists == 1
}

func (s *Service) GetUIPrefs(id string) string {
	var prefs string
	if err := s.db.QueryRow(`SELECT ui_prefs FROM users WHERE id = ?`, id).Scan(&prefs); err != nil || prefs == "" {
		return "{}"
	}
	return prefs
}

func (s *Service) SetUIPrefs(id, prefs string) error {
	_, err := s.db.Exec(`UPDATE users SET ui_prefs = ? WHERE id = ?`, prefs, id)
	return err
}
