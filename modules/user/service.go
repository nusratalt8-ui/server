package user

import "agentmanager/modules/dbutil"

type Service struct {
	db *dbutil.DB
}

func NewService(db *dbutil.DB) *Service {
	return &Service{db: db}
}

type UserWithPlan struct {
	User
	Plan int `json:"plan"`
}

func (s *Service) ListAll() ([]UserWithPlan, error) {
	rows, err := s.db.Query(`SELECT id, username, created_at, plan FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []UserWithPlan
	for rows.Next() {
		var u UserWithPlan
		if err := rows.Scan(&u.ID, &u.Username, &u.CreatedAt, &u.Plan); err != nil {
			continue
		}
		out = append(out, u)
	}
	return out, nil
}
