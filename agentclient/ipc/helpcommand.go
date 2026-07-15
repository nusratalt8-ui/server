package ipc

import (
	"sort"
	"strings"
)

func (cm *Commands) names() []string {
	cm.mu.Lock()
	out := make([]string, 0, len(cm.handlers))
	for n := range cm.handlers {
		out = append(out, n)
	}
	cm.mu.Unlock()
	sort.Strings(out)
	return out
}

func (cm *Commands) help(c *Context) error {
	cm.mu.Lock()
	descs := make(map[string]string, len(cm.handlers))
	cats := make(map[string]string, len(cm.handlers))
	for n, cmd := range cm.handlers {
		descs[n] = cmd.desc
		cats[n] = cmd.cat
	}
	cm.mu.Unlock()
	grouped := make(map[string][]string)
	for _, n := range cm.names() {
		if n == "help" {
			continue
		}
		cat := cats[n]
		if cat == "" {
			cat = "general"
		}
		grouped[cat] = append(grouped[cat], "- `"+cm.prefix+n+"` — "+descs[n])
	}
	catNames := make([]string, 0, len(grouped))
	for cat := range grouped {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	e := NewEmbed().
		WithTitle("Available Commands").
		WithDescription("Use `" + cm.prefix + "<command>` to execute. Type `" + cm.prefix + "<command>` with no args for usage info.").
		WithColor("#3a5f8a")
	for _, cat := range catNames {
		e.AddField(cat, strings.Join(grouped[cat], "\n"), false)
	}
	e.WithFooter("Run " + cm.prefix + "help to see this")
	return c.ReplyEmbed(e)
}
