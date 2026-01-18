package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// ErrSelectionCancelled is returned when the user cancels a selection.
var ErrSelectionCancelled = errors.New("selection cancelled")

// PromptYesNo prompts the user for a yes/no response using stdin/stdout.
func PromptYesNo(prompt string) bool {
	return PromptYesNoWithReader(prompt, os.Stdin, os.Stdout)
}

// PromptYesNoWithReader prompts for yes/no with custom reader/writer for testing.
func PromptYesNoWithReader(prompt string, reader io.Reader, writer io.Writer) bool {
	scanner := bufio.NewScanner(reader)

	for {
		_, _ = fmt.Fprintf(writer, "%s (y/n): ", prompt)
		if !scanner.Scan() {
			return false
		}

		input := strings.TrimSpace(strings.ToLower(scanner.Text()))

		switch input {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
		// Invalid input, loop continues
	}
}

// PromptSelection displays a list and prompts the user to select an item.
// Returns 0-based index of selected item or error if cancelled (user enters 0).
func PromptSelection[T any](items []T, prompt string, display func(index int, item T)) (int, error) {
	return PromptSelectionWithReader(items, prompt, os.Stdin, os.Stdout, display)
}

// PromptSelectionWithReader is the testable version of PromptSelection.
func PromptSelectionWithReader[T any](items []T, prompt string, reader io.Reader, writer io.Writer, display func(index int, item T)) (int, error) {
	// Display items
	for i, item := range items {
		display(i, item)
	}

	scanner := bufio.NewScanner(reader)

	for {
		_, _ = fmt.Fprintf(writer, "%s (0 to cancel): ", prompt)
		if !scanner.Scan() {
			return -1, ErrSelectionCancelled
		}

		input := strings.TrimSpace(scanner.Text())
		num, err := strconv.Atoi(input)
		if err != nil {
			_, _ = fmt.Fprintln(writer, "Please enter a number")
			continue
		}

		if num == 0 {
			return -1, ErrSelectionCancelled
		}

		if num < 1 || num > len(items) {
			_, _ = fmt.Fprintf(writer, "Please enter a number between 1 and %d\n", len(items))
			continue
		}

		return num - 1, nil // Convert to 0-based index
	}
}

// PromptConfirmation prompts for confirmation with a yes/no question.
func PromptConfirmation(prompt string) (bool, error) {
	return PromptConfirmationWithReader(prompt, os.Stdin, os.Stdout)
}

// PromptConfirmationWithReader is the testable version of PromptConfirmation.
func PromptConfirmationWithReader(prompt string, reader io.Reader, writer io.Writer) (bool, error) {
	result := PromptYesNoWithReader(prompt, reader, writer)
	return result, nil
}

// ReadInt reads an integer from stdin.
func ReadInt() (int, error) {
	return ReadIntWithReader(os.Stdin)
}

// ReadIntWithReader reads an integer from a reader.
func ReadIntWithReader(reader io.Reader) (int, error) {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return 0, errors.New("no input")
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return 0, errors.New("empty input")
	}

	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("invalid integer: %s", input)
	}

	return num, nil
}

// ReadString reads a string from stdin with trimming.
func ReadString() (string, error) {
	return ReadStringWithReader(os.Stdin)
}

// ReadStringWithReader reads a trimmed string from a reader.
func ReadStringWithReader(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	if !scanner.Scan() {
		return "", errors.New("no input")
	}

	return strings.TrimSpace(scanner.Text()), nil
}
