package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func PromptInt(prompt string, defaultValue int, minValue int, maxValue int) (int, error) {
	for {
		fmt.Printf("%s [%d-%d]: ", prompt, minValue, maxValue)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return defaultValue, nil
		}

		value, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid number format. Please enter a number between %d and %d.\n", minValue, maxValue)
			continue
		}

		if value < minValue || value > maxValue {
			fmt.Printf("Value must be between %d and %d.\n", minValue, maxValue)
			continue
		}

		return value, nil
	}
}

func PromptString(prompt string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

func PromptPassword(prompt string) string {
	fmt.Printf("%s: ", prompt)
	password, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return ""
	}
	return string(password)
}

func PromptYesNo(prompt string, defaultYes bool) bool {
	defaultStr := "y"
	if !defaultYes {
		defaultStr = "n"
	}

	// Gunakan fmt.Printf untuk memastikan tidak ada buffer yang tersisa
	fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch strings.ToLower(input) {
	case "y", "yes":
		return true
	case "n", "no":
		return false
	case "":
		return defaultYes
	default:
		fmt.Println("Please enter 'y' for yes or 'n' for no.")
		return PromptYesNo(prompt, defaultYes)
	}
}
