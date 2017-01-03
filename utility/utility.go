package utility

import "strings"

// ObfuscateEmail accepts an email address and returns parts of it obfuscated with asterisks.
func ObfuscateEmail(address string) string {
	s := strings.Split(address, "@")

	// There's no ampersand present, more than one, or nothing preceding it, so we don't have a valid
	// local part.
	if len(s) == 1 || len(s) > 2 || len(s[0]) == 0 {
		return address
	}

	lp := string(s[0][0]) + "***"
	domain := strings.Split(s[1], ".")

	// There's no period present or more than one, or it does not have at least one character before
	// the dot, so we don't have a valid domain.
	if len(domain) == 1 || len(domain) > 2 || len(domain[0]) == 0 {
		return address
	}

	d := string(domain[0][0]) + "***"

	return lp + "@" + d + "." + domain[1]
}
