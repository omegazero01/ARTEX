//go:build windows

package server

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

// resolveRealPath returns the final path of p with every reparse point (symlink,
// junction, mount point) resolved. filepath.EvalSymlinks deliberately leaves
// junctions alone on Windows, so the handle-based API is used instead.
func resolveRealPath(p string) (string, bool) {
	p = filepath.Clean(p)
	ptr, err := windows.UTF16PtrFromString(p)
	if err != nil {
		return "", false
	}
	h, err := windows.CreateFile(ptr, 0,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE,
		nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return "", false
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, 512)
	for {
		n, err := windows.GetFinalPathNameByHandle(h, &buf[0], uint32(len(buf)), 0)
		if err != nil {
			return "", false
		}
		if int(n) > len(buf) {
			buf = make([]uint16, n+1)
			continue
		}
		s := windows.UTF16ToString(buf[:n])
		switch {
		case strings.HasPrefix(s, `\\?\UNC\`):
			s = `\\` + strings.TrimPrefix(s, `\\?\UNC\`)
		case strings.HasPrefix(s, `\\?\`):
			s = strings.TrimPrefix(s, `\\?\`)
		}
		return s, true
	}
}

// pathInside reports whether child is base or below it. Windows paths are
// case-insensitive.
func pathInside(child, base string) bool {
	child, base = strings.ToLower(child), strings.ToLower(base)
	return child == base || strings.HasPrefix(child, base+`\`)
}
