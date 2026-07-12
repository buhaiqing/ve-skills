package run

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func joinPath(parts ...string) string { return filepath.Join(parts...) }

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

func relPath(base, target string) (string, error) { return filepath.Rel(base, target) }

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func firstN(s []string, n int) []string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func orEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func orString(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func isTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func readAllStdin() ([]byte, error) { return io.ReadAll(os.Stdin) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
