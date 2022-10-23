package email

import (
	"regexp"
	"strings"

	"github.com/go-msvc/errors"
)

const (
	namePattern    = `[a-z0-9-]+`
	namesPattern   = namePattern + `(\.` + namePattern + `)*`
	domainPattern  = `[a-z0-9-]+`
	domainsPattern = domainPattern + `(\.` + domainPattern + `)+`
	emailPattern   = namesPattern + "@" + domainsPattern
)

var (
	emailRegex     = regexp.MustCompile("^" + emailPattern + "$")
	errInvalidAddr = errors.Errorf("invalid email address")
)

func Valid(emailAddr string) (string, error) {
	s := strings.ToLower(emailAddr)
	if emailRegex.MatchString(s) {
		return s, nil
	}
	return "", errors.Wrapf(errInvalidAddr, "email(%s)", emailAddr)
}
