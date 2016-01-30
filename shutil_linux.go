// +build linux,cgo

package shutil

/*
#include <sys/ioctl.h>

#undef BTRFS_IOCTL_MAGIC
#define BTRFS_IOCTL_MAGIC 0x94
#undef BTRFS_IOC_CLONE
#define BTRFS_IOC_CLONE _IOW (BTRFS_IOCTL_MAGIC, 9, int)
*/
import "C"

import (
	"os"
	"syscall"
)

const (
	BtrfsIocClone = C.BTRFS_IOC_CLONE
)

func clonefile(fdst *os.File, fsrc *os.File) (bool, error) {
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, fdst.Fd(), BtrfsIocClone, fsrc.Fd()); err != 0 {
		return false, err
	}
	return true, nil
}
