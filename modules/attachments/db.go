package attachments

type Record struct {
	ID          string
	Filename    string
	ContentType string
	Size        int64
	CreatedAt   int64
}

func (s *Service) insert(r Record) error {
	_, err := s.db.Exec(
		`INSERT INTO attachments (id, filename, content_type, size, created_at) VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.Filename, r.ContentType, r.Size, r.CreatedAt,
	)
	return err
}

func (s *Service) get(id string) (Record, bool) {
	var r Record
	err := s.db.QueryRow(
		`SELECT id, filename, content_type, size, created_at FROM attachments WHERE id = ?`, id,
	).Scan(&r.ID, &r.Filename, &r.ContentType, &r.Size, &r.CreatedAt)
	if err != nil {
		return Record{}, false
	}
	return r, true
}
