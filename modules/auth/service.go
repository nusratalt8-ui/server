package auth

import (
	"time"

	"agentmanager/modules/dbutil"
	"agentmanager/modules/plan"
	"agentmanager/modules/user"
)

const ctxUserID = "auth_user_id"

type Service struct {
	db         *dbutil.DB
	cfg        Config
	user       *user.Service
	plan       *plan.Service
	onRegister func(userID string) error
}

func NewService(db *dbutil.DB, cfg Config, userSvc *user.Service, planSvc *plan.Service) *Service {
	s := &Service{db: db, cfg: cfg, user: userSvc, plan: planSvc}
	go s.reapLoop()
	return s
}

func (s *Service) SetOnRegister(fn func(userID string) error) {
	s.onRegister = fn
}

func (s *Service) SetPlan(planSvc *plan.Service) {
	s.plan = planSvc
}

func (s *Service) reapLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		s.deleteExpiredSessions(time.Now().Unix())
	}
}
