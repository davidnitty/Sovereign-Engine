package main

import (
	"fmt"
	"os"

	"github.com/yourusername/the-engine/internal/app"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: engine-encrypt encrypt|decrypt <string>")
		os.Exit(2)
	}
	a, err := app.New(envOrDefault("ENGINE_COMPOSITION_DIR", "compositions"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer a.Close()
	switch os.Args[1] {
	case "encrypt":
		out, err := a.Encryption.Encrypt(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(out)
	case "decrypt":
		out, err := a.Encryption.Decrypt(os.Args[2])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(out)
	default:
		fmt.Fprintln(os.Stderr, "usage: engine-encrypt encrypt|decrypt <string>")
		os.Exit(2)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
