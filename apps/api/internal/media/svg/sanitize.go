package svg

import (
	"bytes"
	"errors"
	"regexp"
)

var (
	scriptTagPattern = regexp.MustCompile(`(?is)<\s*script[\s>].*?<\s*/\s*script\s*>`)
	eventAttrPattern = regexp.MustCompile(`(?is)\son[a-z]+\s*=\s*"[^"]*"`)
)

func Sanitize(input []byte) ([]byte, error) {
	if !bytes.Contains(bytes.ToLower(input), []byte("<svg")) {
		return nil, errors.New("not an svg document")
	}

	clean := scriptTagPattern.ReplaceAll(input, nil)
	clean = eventAttrPattern.ReplaceAll(clean, nil)

	return clean, nil
}
