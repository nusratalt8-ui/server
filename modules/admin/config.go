package admins

var Usernames = []string{
	"Admin1",
	"vectorone",
	"DrDonutt",
}

var resolvedIDs = map[string]struct{}{}

func Resolve(lookup func(username string) (id string, ok bool)) {
	for _, u := range Usernames {
		if id, ok := lookup(u); ok {
			resolvedIDs[id] = struct{}{}
		}
	}
}

func IsAdmin(userID string) bool {
	_, ok := resolvedIDs[userID]
	return ok
}

func AdminIDs() []string {
	out := make([]string, 0, len(resolvedIDs))
	for id := range resolvedIDs {
		out = append(out, id)
	}
	return out
}
