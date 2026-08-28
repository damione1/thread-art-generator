package pbErrors

import (
	"strings"
	"unicode"
)

const formGlobalKey = "_form"

// FormFields maps a Connect error onto the key space templates index:
// proto path (art.title), last segment (title), snake/camel, and _form.
func FormFields(err error) map[string][]string {
	standardErr := FromConnectError(err)
	if standardErr == nil {
		return map[string][]string{}
	}

	out := ExpandFieldKeys(standardErr.Fields)
	if standardErr.GlobalError != "" {
		out[formGlobalKey] = []string{standardErr.GlobalError}
	}
	if len(out) == 0 && standardErr.Message != "" {
		msg := standardErr.Message
		if standardErr.Type == ErrorTypeInternal {
			msg = UserFacingInternalMessage
		}
		out[formGlobalKey] = []string{msg}
	}
	return out
}

// ExpandFieldKeys copies proto paths and adds last-segment + snake/camel aliases.
func ExpandFieldKeys(fields map[string][]string) map[string][]string {
	out := make(map[string][]string)
	for field, messages := range fields {
		if field == "" {
			continue
		}
		for _, alias := range FieldAliases(field) {
			out[alias] = appendUnique(out[alias], messages)
		}
	}
	return out
}

// FieldMessages returns errors for a form field name or proto path.
func FieldMessages(fields map[string][]string, name string) []string {
	if len(fields) == 0 || name == "" {
		return nil
	}
	if msgs, ok := fields[name]; ok && len(msgs) > 0 {
		return msgs
	}
	want := FieldAliases(name)
	seen := map[string]struct{}{}
	var out []string
	for key, messages := range fields {
		if key == formGlobalKey {
			continue
		}
		aliases := FieldAliases(key)
		if !overlaps(want, aliases) {
			continue
		}
		for _, msg := range messages {
			if _, dup := seen[msg]; dup {
				continue
			}
			seen[msg] = struct{}{}
			out = append(out, msg)
		}
	}
	return out
}

func FieldAliases(field string) []string {
	if field == "" {
		return nil
	}
	aliases := []string{field}
	if i := strings.LastIndex(field, "."); i >= 0 && i < len(field)-1 {
		aliases = append(aliases, field[i+1:])
	}
	var extra []string
	for _, a := range aliases {
		if snake := camelToSnake(a); snake != a {
			extra = append(extra, snake)
		}
		if camel := snakeToCamel(a); camel != a {
			extra = append(extra, camel)
		}
	}
	aliases = append(aliases, extra...)
	return unique(aliases)
}

func overlaps(a, b []string) bool {
	set := make(map[string]struct{}, len(a))
	for _, x := range a {
		set[x] = struct{}{}
	}
	for _, y := range b {
		if _, ok := set[y]; ok {
			return true
		}
	}
	return false
}

func unique(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func appendUnique(dst, src []string) []string {
	seen := make(map[string]struct{}, len(dst))
	for _, s := range dst {
		seen[s] = struct{}{}
	}
	for _, s := range src {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		dst = append(dst, s)
	}
	return dst
}

func snakeToCamel(s string) string {
	if !strings.Contains(s, "_") {
		return s
	}
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		r := []rune(parts[i])
		r[0] = unicode.ToUpper(r[0])
		parts[i] = string(r)
	}
	return strings.Join(parts, "")
}

func camelToSnake(s string) string {
	if s == "" || strings.Contains(s, "_") {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
