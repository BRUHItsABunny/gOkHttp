package constants

import "errors"

var (
	ErrHostNotFound           = errors.New("host not found")
	ErrUnmatchedAlgo          = errors.New("unmatched algorithm")
	ErrCookieJarNotPersistent = errors.New("cookie jar is not persistent")
)
