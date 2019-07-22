package cookies

import (
	"fmt"
	"strings"
)

// id returns the domain;path;name triple of e as an id.
func (e *entry) id() string {
	return id(e.Domain, e.Path, e.Name)
}

// id returns the domain;path;name triple as an id.
func id(domain, path, name string) string {
	return fmt.Sprintf("%s;%s;%s", domain, path, name)
}

// shouldSend determines whether e's cookie qualifies to be included in a
// request to host/path. It is the caller's responsibility to check if the
// cookie is expired.
func (e *entry) shouldSend(https bool, host, path string) bool {
	return e.domainMatch(host) && e.pathMatch(path) && (https || !e.Secure)
}

// domainMatch implements "domain-match" of RFC 6265 section 5.1.3.
func (e *entry) domainMatch(host string) bool {
	if e.Domain == host {
		return true
	}
	return !e.HostOnly && hasDotSuffix(host, e.Domain)
}

// pathMatch implements "path-match" according to RFC 6265 section 5.1.4.
func (e *entry) pathMatch(requestPath string) bool {
	if requestPath == e.Path {
		return true
	}
	if strings.HasPrefix(requestPath, e.Path) {
		if e.Path[len(e.Path)-1] == '/' {
			return true // The "/any/" matches "/any/path" case.
		} else if requestPath[len(e.Path)] == '/' {
			return true // The "/any" matches "/any/path" case.
		}
	}
	return false
}
