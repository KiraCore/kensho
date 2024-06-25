package utils

import (
	"fmt"

	"github.com/atotto/clipboard"
)

func CopyToClipboard(text string) error {
	// Copy the provided text to the clipboard
	err := clipboard.WriteAll(text)
	if err != nil {
		return fmt.Errorf("failed to copy text to clipboard: %w", err)
	}
	return nil
}
