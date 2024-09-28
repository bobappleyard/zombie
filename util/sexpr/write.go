package sexpr

import "strings"

func WriteString(e Expr) string {
	if e.Kind() != List {
		return e.Text()
	}
	var sb strings.Builder
	sb.WriteByte('(')
	more := false
	for !e.Empty() {
		if more {
			sb.WriteByte(' ')
		}
		more = true
		sb.WriteString(WriteString(e.Head()))
		e = e.Tail()
	}
	sb.WriteByte(')')
	return sb.String()
}
