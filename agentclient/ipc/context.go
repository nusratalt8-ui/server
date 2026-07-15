package ipc

import "strings"

type Handler func(c *Context) error

type Context struct {
	Name        string
	Args        []string
	Attachments []string
	Raw         string
	prefix      string
	send        func(text string, embed *Embed, attachments []string) error
}

func (c *Context) Reply(text string) error      { return c.send(text, nil, nil) }
func (c *Context) ReplyEmbed(e *Embed) error    { return c.send("", e, nil) }
func (c *Context) ReplyFile(ids []string) error { return c.send("", nil, ids) }
func (c *Context) ArgString() string            { return strings.Join(c.Args, " ") }
func (c *Context) Prefix() string               { return c.prefix }

func (c *Context) Usage(lines []string) error {
	e := NewEmbed().WithTitle(c.prefix + c.Name).WithColor("#a33")
	for _, line := range lines {
		e.AddField("", "• "+line, false)
	}
	return c.ReplyEmbed(e)
}
