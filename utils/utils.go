package utils

import (
	"fmt"
	"os"
	"path/filepath"

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

func InitHomeFolder(appName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	appHome := filepath.Join(homeDir, "."+appName)

	err = os.MkdirAll(appHome, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create app home directory: %w", err)
	}

	return appHome, nil
}
