package agentkey

import "agentmanager/modules/dbutil"

type Service struct {
	db       *dbutil.DB
	onRotate func(userID string)
}

func NewService(db *dbutil.DB) *Service {
	return &Service{db: db}
}

func (s *Service) SetOnRotate(fn func(userID string)) {
	s.onRotate = fn
}
