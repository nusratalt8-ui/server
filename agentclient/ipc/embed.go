package ipc

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type Embed struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Color       string  `json:"color,omitempty"`
	Fields      []Field `json:"fields,omitempty"`
	Footer      string  `json:"footer,omitempty"`
	Thumbnail   string  `json:"thumbnail,omitempty"`
}

func NewEmbed() *Embed { return &Embed{} }

func Success(msg string) *Embed {
	return &Embed{Title: "✅ Success", Description: msg, Color: "#2d7a2d"}
}

func Error(msg string) *Embed {
	return &Embed{Title: "❌ Error", Description: msg, Color: "#a33"}
}

func (e *Embed) WithTitle(t string) *Embed       { e.Title = t; return e }
func (e *Embed) WithDescription(d string) *Embed { e.Description = d; return e }
func (e *Embed) WithColor(c string) *Embed       { e.Color = c; return e }
func (e *Embed) WithFooter(f string) *Embed      { e.Footer = f; return e }
func (e *Embed) WithThumbnail(url string) *Embed { e.Thumbnail = url; return e }
func (e *Embed) AddField(name, value string, inline bool) *Embed {
	e.Fields = append(e.Fields, Field{Name: name, Value: value, Inline: inline})
	return e
}
