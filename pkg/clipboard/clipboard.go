package clipboard

import (
	"errors"
	"io"
	"os/exec"
)

// Copy writes text to the system clipboard (supporting Wayland and X11).
func Copy(text string) error {
	// 1. Wayland wl-copy
	if path, err := exec.LookPath("wl-copy"); err == nil {
		cmd := exec.Command(path)
		stdin, err := cmd.StdinPipe()
		if err == nil {
			go func() {
				defer stdin.Close()
				_, _ = io.WriteString(stdin, text)
			}()
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	// 2. X11 xclip
	if path, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command(path, "-selection", "clipboard")
		stdin, err := cmd.StdinPipe()
		if err == nil {
			go func() {
				defer stdin.Close()
				_, _ = io.WriteString(stdin, text)
			}()
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	// 3. X11 xsel
	if path, err := exec.LookPath("xsel"); err == nil {
		cmd := exec.Command(path, "--clipboard", "--input")
		stdin, err := cmd.StdinPipe()
		if err == nil {
			go func() {
				defer stdin.Close()
				_, _ = io.WriteString(stdin, text)
			}()
			if err := cmd.Run(); err == nil {
				return nil
			}
		}
	}

	return errors.New("clipboard utility not found (please install wl-copy or xclip)")
}
