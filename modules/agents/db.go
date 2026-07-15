package agents

func (s *Service) upsert(a Agent, now int64) error {
	_, err := s.db.Exec(
		`INSERT INTO agents (id, user_id, name, description, hostname, username, online, last_seen)
		 VALUES (?, ?, ?, ?, ?, ?, 1, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   name = excluded.name,
		   description = excluded.description,
		   hostname = excluded.hostname,
		   username = excluded.username,
		   online = 1,
		   last_seen = excluded.last_seen`,
		a.ID, a.UserID, a.Name, a.Description, a.Hostname, a.Username, now,
	)
	return err
}

func (s *Service) setOffline(id string, now int64) error {
	_, err := s.db.Exec(`UPDATE agents SET online = 0, last_seen = ? WHERE id = ?`, now, id)
	return err
}

func (s *Service) exists(id, userID string) bool {
	var one int
	err := s.db.QueryRow(`SELECT 1 FROM agents WHERE id = ? AND user_id = ?`, id, userID).Scan(&one)
	return err == nil
}

func (s *Service) resetAllOffline() error {
	_, err := s.db.Exec(`UPDATE agents SET online = 0`)
	return err
}

func (s *Service) ownerOf(id string) (string, bool) {
	var userID string
	if err := s.db.QueryRow(`SELECT user_id FROM agents WHERE id = ?`, id).Scan(&userID); err != nil {
		return "", false
	}
	return userID, true
}

func (s *Service) getByID(id string) (Agent, bool) {
	var a Agent
	var online int
	err := s.db.QueryRow(
		`SELECT a.id, a.user_id, u.username, a.name, a.description, a.hostname, a.username, a.online, a.last_seen
		 FROM agents a LEFT JOIN users u ON u.id = a.user_id WHERE a.id = ?`, id,
	).Scan(&a.ID, &a.UserID, &a.OwnerUsername, &a.Name, &a.Description, &a.Hostname, &a.Username, &online, &a.LastSeen)
	if err != nil {
		return Agent{}, false
	}
	a.Online = online == 1
	return a, true
}

func (s *Service) countsForUser(userID string) (total, online int, err error) {
	err = s.db.QueryRow(
		`SELECT COUNT(*), SUM(CASE WHEN online=1 THEN 1 ELSE 0 END)
		 FROM agents WHERE user_id = ?`, userID,
	).Scan(&total, &online)
	return
}

func (s *Service) countsPerUser() (map[string]UserCounts, int, int, error) {
	rows, err := s.db.Query(
		`SELECT a.user_id,
		        COUNT(*) AS total,
		        SUM(CASE WHEN a.online=1 THEN 1 ELSE 0 END) AS online_count,
		        u.username
		 FROM agents a
		 LEFT JOIN users u ON u.id = a.user_id
		 GROUP BY a.user_id`)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	out := make(map[string]UserCounts)
	var totalAll, onlineAll int
	for rows.Next() {
		var uid string
		var c UserCounts
		if err := rows.Scan(&uid, &c.Total, &c.Online, &c.Username); err != nil {
			return nil, 0, 0, err
		}
		out[uid] = c
		totalAll += c.Total
		onlineAll += c.Online
	}
	return out, totalAll, onlineAll, rows.Err()
}

func (s *Service) listPaged(userID string, limit, offset int) ([]Agent, error) {
	rows, err := s.db.Query(
		`SELECT id, user_id, name, description, hostname, username, online, last_seen
		 FROM agents WHERE user_id = ?
		 ORDER BY online DESC, last_seen DESC
		 LIMIT ? OFFSET ?`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Agent, 0, limit)
	for rows.Next() {
		var a Agent
		var online int
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Description, &a.Hostname, &a.Username, &online, &a.LastSeen); err != nil {
			return nil, err
		}
		a.Online = online == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Service) listByUserPaged(userID string, limit, offset int) ([]Agent, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.user_id, u.username, a.name, a.description, a.hostname, a.username, a.online, a.last_seen
		 FROM agents a
		 LEFT JOIN users u ON u.id = a.user_id
		 WHERE a.user_id = ?
		 ORDER BY a.online DESC, a.last_seen DESC
		 LIMIT ? OFFSET ?`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Agent, 0, limit)
	for rows.Next() {
		var a Agent
		var online int
		if err := rows.Scan(&a.ID, &a.UserID, &a.OwnerUsername, &a.Name, &a.Description, &a.Hostname, &a.Username, &online, &a.LastSeen); err != nil {
			return nil, err
		}
		a.Online = online == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Service) deleteAgent(id string) error {
	_, err := s.db.Exec(`DELETE FROM agents WHERE id = ?`, id)
	return err
}

func (s *Service) listAllPaged(limit, offset int) ([]Agent, error) {
	rows, err := s.db.Query(
		`SELECT a.id, a.user_id, u.username, a.name, a.description, a.hostname, a.username, a.online, a.last_seen
		 FROM agents a
		 LEFT JOIN users u ON u.id = a.user_id
		 ORDER BY a.online DESC, a.last_seen DESC
		 LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Agent, 0, limit)
	for rows.Next() {
		var a Agent
		var online int
		if err := rows.Scan(&a.ID, &a.UserID, &a.OwnerUsername, &a.Name, &a.Description, &a.Hostname, &a.Username, &online, &a.LastSeen); err != nil {
			return nil, err
		}
		a.Online = online == 1
		out = append(out, a)
	}
	return out, rows.Err()
}
