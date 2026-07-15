package plan

import "agentmanager/modules/admin"

const (
	PlanFree = 0
	PlanPaid = 1
)

func (s *Service) Get(userID string) int {
	var plan int
	if err := s.db.QueryRow(`SELECT plan FROM users WHERE id = ?`, userID).Scan(&plan); err != nil {
		return PlanFree
	}
	return plan
}

func (s *Service) Set(userID string, plan int) error {
	_, err := s.db.Exec(`UPDATE users SET plan = ? WHERE id = ?`, plan, userID)
	return err
}

func (s *Service) IsPaid(userID string) bool {
	if s.Get(userID) >= PlanPaid {
		return true
	}
	return admins.IsAdmin(userID)
}
