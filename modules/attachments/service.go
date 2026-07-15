package attachments

import (
	"os"

	"agentmanager/modules/dbutil"
)

type Service struct {
	db  *dbutil.DB
	cfg Config
}

func NewService(db *dbutil.DB, cfg Config) (*Service, error) {
	if err := os.MkdirAll(cfg.StoragePath, 0o755); err != nil {
		return nil, err
	}
	return &Service{db: db, cfg: cfg}, nil
}

func (s *Service) Exists(id string) bool {
	_, ok := s.get(id)
	return ok
}

func (s *Service) Meta(id string) (filename, contentType string, size int64, ok bool) {
	r, found := s.get(id)
	if !found {
		return "", "", 0, false
	}
	return r.Filename, r.ContentType, r.Size, true
}
