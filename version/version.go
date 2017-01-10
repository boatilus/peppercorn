package version

import "fmt"

// Follows semantic versioning: http://semver.org/
const (
	versionMajor = 0
	versionMinor = 1
	versionPatch = 2
)

// GetString returns the version as a string.
func GetString() string {
	return fmt.Sprintf("%d.%d.%d", versionMajor, versionMinor, versionPatch)
}
