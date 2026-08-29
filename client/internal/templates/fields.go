package templates

import pbErrors "github.com/Damione1/thread-art-generator/core/errors"

// FieldMsgs looks up form errors by proto path or HTML name (art.title / title).
func FieldMsgs(errs map[string][]string, field string) []string {
	return pbErrors.FieldMessages(errs, field)
}

func HasFieldErr(errs map[string][]string, field string) bool {
	return len(FieldMsgs(errs, field)) > 0
}
