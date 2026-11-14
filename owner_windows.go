package main

import (
	"os/user"

	"golang.org/x/sys/windows"
)

// fileOwner returns the full name of the owner of the file, or the username if
// the full name is not known.
func fileOwner(filename string) (string, error) {
	descriptor, err := windows.GetNamedSecurityInfo(filename, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION)
	if err != nil {
		return "", err
	}
	owner, _, err := descriptor.Owner()
	if err != nil {
		return "", err
	}
	username, _, _, err := owner.LookupAccount("")
	if err != nil {
		return "", err
	}
	// The full name is only available for user accounts, not for groups like
	// the administrators. For an account without a full name the lookup
	// returns the username with the domain instead.
	if account, err := user.LookupId(owner.String()); err == nil && account.Name != "" && account.Name != account.Username {
		return account.Name, nil
	}
	return username, nil
}
