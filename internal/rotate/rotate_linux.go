package rotate

import (
	"os"
	"syscall"
)

// inoOf extracts the inode number from a FileInfo on Linux.
func inoOf(info os.FileInfo) uint64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Ino
	}
	return 0
}
