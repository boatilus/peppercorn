package users

import "github.com/boatilus/peppercorn/utility"

// GenerateRecoveryCodes creates a set of 10 new MFA recovery codes on the user.
func (u *User) GenerateRecoveryCodes() {
	u.RecoveryCodes = make([]string, 10)
	for i := 0; i < 10; i++ {
		u.RecoveryCodes[i] = utility.GenerateRandomRecoveryCode()
	}
}
