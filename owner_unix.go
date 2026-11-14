//go:build unix

package main

import (
	"cmp"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// fileOwner returns the full name of the owner of the file, or the username if
// the full name is not known.
func fileOwner(filename string) (string, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return "", err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return "", fmt.Errorf("%s: owner not available", filename)
	}
	owner, err := user.LookupId(strconv.FormatUint(uint64(stat.Uid), 10))
	if err != nil {
		return "", err
	}
	return cmp.Or(owner.Name, owner.Username), nil
}
