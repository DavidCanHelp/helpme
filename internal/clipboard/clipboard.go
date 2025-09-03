package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy copies text to the system clipboard
func Copy(text string) error {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, then xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("clipboard: no clipboard utility found (install xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("cmd", "/c", "clip")
	default:
		return fmt.Errorf("clipboard: unsupported platform %s", runtime.GOOS)
	}
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	
	if _, err := stdin.Write([]byte(text)); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("clipboard: %w", err)
	}
	
	return cmd.Wait()
}

// CopyWithFeedback copies text to clipboard and returns a user-friendly message
func CopyWithFeedback(text, description string) string {
	if err := Copy(text); err != nil {
		// Fallback: show the text for manual copying
		return fmt.Sprintf("Could not copy to clipboard. Please copy manually: %s", text)
	}
	return fmt.Sprintf("✓ Copied %s to clipboard: %s", description, text)
}

// FormatResourceForCopy formats a resource's contact info for clipboard
func FormatResourceForCopy(name, number, text, sms string) string {
	var parts []string
	parts = append(parts, name)
	if number != "" {
		parts = append(parts, "Phone: "+number)
	}
	if text != "" {
		parts = append(parts, "Text: "+text)
	}
	if sms != "" {
		parts = append(parts, "SMS: "+sms)
	}
	return strings.Join(parts, "\n")
}