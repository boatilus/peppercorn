package utility

import "strings"

// ObfuscateEmail accepts an email address and returns parts of it obfuscated with asterisks.
func ObfuscateEmail(address string) string {
	s := strings.Split(address, "@")

	if len(s) == 1 {
		return address
	}

	if len(s[0]) == 1 {
		s[0] = s[0] + "**"
	} else {
		s[0] = string(s[0][0]) + string(s[0][1]) + "***"
	}

	domain := strings.Split(s[1], ".")

	if len(domain) == 1 {
		return address
	}

	if len(domain[0]) == 0 {
		return address
	}

	domain[0] = string(domain[0][0]) + "***"

	return s[0] + "@" + domain[0] + "." + domain[1]
}
