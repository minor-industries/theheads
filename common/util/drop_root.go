package util

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

func DropRoot(
	logger *zap.Logger,
	userName string,
) error {
	if os.Geteuid() != 0 {
		logger.Info("already running with non-root privileges")
		return nil
	}

	logger.Info("running as root, dropping privileges")

	nonRootUser, err := user.Lookup(userName)
	if err != nil {
		return errors.Wrap(err, "lookup user")
	}

	// change the group first, since we won't be able to after changing the user
	gid, err := strconv.Atoi(nonRootUser.Gid)
	if err != nil {
		return errors.Wrap(err, "group atoi")
	}

	if err := syscall.Setgid(gid); err != nil {
		return errors.Wrap(err, "set gid")
	}

	uid, err := strconv.Atoi(nonRootUser.Uid)
	if err != nil {
		return errors.Wrap(err, "user atoi")
	}

	if err := syscall.Setuid(uid); err != nil {
		return errors.Wrap(err, "set uid")
	}

	logger.Info("privileges dropped")
	return nil
}
