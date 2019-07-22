package cookies

import (
	"crypto/cipher"
	"sync"
	"time"
)

type Crypt struct {
	Cipher cipher.Block
}

// PublicSuffixList provides the public suffix of a domain. For example:
//      - the public suffix of "example.com" is "com",
//      - the public suffix of "foo1.foo2.foo3.co.uk" is "co.uk", and
//      - the public suffix of "bar.pvt.k12.ma.us" is "pvt.k12.ma.us".
//
// Implementations of PublicSuffixList must be safe for concurrent use by
// multiple goroutines.
//
// An implementation that always returns "" is valid and may be useful for
// testing but it is not secure: it means that the HTTP server for foo.com can
// set a cookie for bar.com.
//
// A public suffix list implementation is in the package
// golang.org/x/net/publicsuffix.
type PublicSuffixList interface {
	// PublicSuffix returns the public suffix of domain.
	//
	// TODO: specify which of the caller and callee is responsible for IP
	// addresses, for leading and trailing dots, for case sensitivity, and
	// for IDN/Punycode.
	PublicSuffix(domain string) string

	// String returns a description of the source of this public suffix
	// list. The description will typically contain something like a time
	// stamp or version number.
	String() string
}

// Options are the options for creating a new Jar.
type JarOptions struct {
	// PublicSuffixList is the public suffix list that determines whether
	// an HTTP server can set a cookie for a domain.
	//
	// If this is nil, the public suffix list implementation in golang.org/x/net/publicsuffix
	// is used.
	PublicSuffixList PublicSuffixList

	// Filename holds the file to use for storage of the cookies.
	// If it is empty, the value of DefaultCookieFile will be used.
	Filename string

	// NoPersist specifies whether no persistence should be used
	// (useful for tests). If this is true, the value of Filename will be
	// ignored.
	NoPersist bool

	/*
		EncryptPassword specifies the password for encrypting the cookie file, if it isn't equal to "" encryption will be used
	*/
	EncryptionPassword string
}

// Jar implements the http.CookieJar interface from the net/http package.
type Jar struct {
	// filename holds the file that the cookies were loaded from.
	filename string

	psList PublicSuffixList

	crypt Crypt

	// mu locks the remaining fields.
	mu sync.Mutex

	// entries is a set of entries, keyed by their eTLD+1 and subkeyed by
	// their name/domain/path.
	entries map[string]map[string]entry

	options *JarOptions
}

// entry is the internal representation of a cookie.
//
// This struct type is not used outside of this package per se, but the exported
// fields are those of RFC 6265.
// Note that this structure is marshaled to JSON, so backward-compatibility
// should be preserved.
type entry struct {
	Name       string
	Value      string
	Domain     string
	Path       string
	Secure     bool
	HttpOnly   bool
	Persistent bool
	HostOnly   bool
	Expires    time.Time
	Creation   time.Time
	LastAccess time.Time

	// Updated records when the cookie was updated.
	// This is different from creation time because a cookie
	// can be changed without updating the creation time.
	Updated time.Time

	// CanonicalHost stores the original canonical host name
	// that the cookie was associated with. We store this
	// so that even if the public suffix list changes (for example
	// when storing/loading cookies) we can still get the correct
	// jar keys.
	CanonicalHost string
}

type byCanonicalHost struct {
	byPathLength
}

type byPathLength []entry
