package filter

import "fmt"

type errorExpectedEOL struct {
	data any
}

func (e errorExpectedEOL) Error() string {
	return fmt.Sprintf("unexpected token(s): %q; expected end of line", e.data)
}
