package server

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func workspaceTestServer(t *testing.T, root string) *Server {
	t.Helper()
	return &Server{m: &Manager{dir: root}}
}

// makeDirLink creates a directory link. os.Symlink needs SeCreateSymbolicLink
// privilege on Windows, so fall back to an NTFS junction (same reparse-point
// semantics for path resolution, no privilege required).
func makeDirLink(t *testing.T, link, target string) bool {
	t.Helper()
	if err := os.Symlink(target, link); err == nil {
		return true
	}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
		if err == nil {
			return true
		}
		t.Logf("junction failed: %v %s", err, strings.TrimSpace(string(out)))
	}
	return false
}

// A link inside the work dir must not become a door to the rest of the disk.
func TestWSResolveRejectsSymlinkOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if !makeDirLink(t, filepath.Join(root, "escape"), outside) {
		t.Skip("cannot create directory link on this host")
	}
	s := workspaceTestServer(t, root)

	for _, rel := range []string{"escape", "escape/secret.txt", "escape/new-file.txt"} {
		if abs, ok := s.wsResolve(rel); ok {
			t.Fatalf("wsResolve(%q) allowed escaping path: %s", rel, abs)
		}
	}
}

// Links that stay inside the work dir remain usable, and a not-yet-existing file
// below such a link is judged by its nearest existing ancestor.
func TestWSResolveAllowsInternalSymlinkAndNewFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "real", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !makeDirLink(t, filepath.Join(root, "alias"), filepath.Join(root, "real")) {
		t.Skip("cannot create directory link on this host")
	}
	s := workspaceTestServer(t, root)

	for _, rel := range []string{"", "real", "real/nested", "alias", "alias/nested", "alias/new.txt", "brand/new/deep.txt"} {
		if _, ok := s.wsResolve(rel); !ok {
			t.Fatalf("wsResolve(%q) rejected a path inside the root", rel)
		}
	}
}

// ".." is clamped to the root (same semantics as before); it must never escape,
// whether through forward or back slashes.
func TestWSResolveClampsTraversalInsideRoot(t *testing.T) {
	root := t.TempDir()
	s := workspaceTestServer(t, root)
	base := filepath.Clean(root)
	for _, rel := range []string{"../x", "a/../../x", "/../x", "..\\x", "../../../../etc/passwd"} {
		abs, ok := s.wsResolve(rel)
		if !ok {
			continue // rejecting is also acceptable
		}
		if abs != base && !strings.HasPrefix(abs, base+string(os.PathSeparator)) {
			t.Fatalf("wsResolve(%q) escaped the root: %s", rel, abs)
		}
	}
}

// A new path below an escaping link must be rejected as well.
func TestWSResolveRejectsNestedEscape(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if !makeDirLink(t, filepath.Join(root, "escape"), outside) {
		t.Skip("cannot create directory link on this host")
	}
	s := workspaceTestServer(t, root)
	if _, ok := s.wsResolve("escape/sub/dir/new.txt"); ok {
		t.Fatal("wsResolve allowed a new path below an escaping link")
	}
}

func TestWSResolveAcceptsRootAndPlainFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	s := workspaceTestServer(t, root)
	for _, rel := range []string{"", "a.txt"} {
		if _, ok := s.wsResolve(rel); !ok {
			t.Fatalf("wsResolve(%q) rejected a valid path", rel)
		}
	}
}
