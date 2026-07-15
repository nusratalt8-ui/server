package plan

import "agentmanager/modules/dbutil"
import "agentmanager/modules/wsutil"

type Service struct {
	db    *dbutil.DB
	panel *wsutil.Hub
}

func NewService(db *dbutil.DB, panel *wsutil.Hub) *Service {
	return &Service{db: db, panel: panel}
}
