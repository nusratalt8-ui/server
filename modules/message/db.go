package message

import "encoding/json"

type Attachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type Message struct {
	ID          string          `json:"id"`
	AgentID     string          `json:"agent_id"`
	Sender      string          `json:"sender"`
	Body        string          `json:"body"`
	Attachments []Attachment    `json:"attachments"`
	Embed       json.RawMessage `json:"embed,omitempty"`
	CreatedAt   int64           `json:"created_at"`
}

func (s *Service) insert(m Message) error {
	ids := make([]string, len(m.Attachments))
	for i, a := range m.Attachments {
		ids[i] = a.ID
	}
	att, _ := json.Marshal(ids)
	embed := ""
	if len(m.Embed) > 0 {
		embed = string(m.Embed)
	}
	_, err := s.db.Exec(
		`INSERT INTO messages (id, agent_id, sender, body, attachments, embed, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.AgentID, m.Sender, m.Body, string(att), embed, m.CreatedAt,
	)
	return err
}

func scanMessages(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close() error
}, limit int) ([]Message, [][]string, error) {
	out := make([]Message, 0, limit)
	idsByRow := make([][]string, 0, limit)
	for rows.Next() {
		var m Message
		var att, embed string
		if err := rows.Scan(&m.ID, &m.AgentID, &m.Sender, &m.Body, &att, &embed, &m.CreatedAt); err != nil {
			rows.Close()
			return nil, nil, err
		}
		var ids []string
		if json.Unmarshal([]byte(att), &ids) != nil {
			ids = []string{}
		}
		if embed != "" {
			m.Embed = json.RawMessage(embed)
		}
		out = append(out, m)
		idsByRow = append(idsByRow, ids)
	}
	rows.Close()
	return out, idsByRow, rows.Err()
}

func (s *Service) fetchMessages(agentID, before string, limit int) ([]Message, error) {
	var (
		rows interface {
			Next() bool
			Scan(...any) error
			Err() error
			Close() error
		}
		err error
	)
	if before != "" {
		rows, err = s.db.Query(
			`SELECT id, agent_id, sender, body, attachments, embed, created_at
			 FROM messages WHERE agent_id = ? AND created_at < (SELECT created_at FROM messages WHERE id = ?)
			 ORDER BY created_at DESC LIMIT ?`,
			agentID, before, limit,
		)
	} else {
		rows, err = s.db.Query(
			`SELECT id, agent_id, sender, body, attachments, embed, created_at
			 FROM messages WHERE agent_id = ? ORDER BY created_at DESC LIMIT ?`,
			agentID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	out, idsByRow, err := scanMessages(rows, limit)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Attachments = s.hydrate(idsByRow[i])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
