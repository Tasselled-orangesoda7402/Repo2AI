package main

import (
	"fmt"
	"os"

	"github.com/fichil/Repo2AI/internal/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Repo2AI Started")
		fmt.Println("Usage: repo2ai scan <path>")
		return
	}

	command := os.Args[1]

	switch command {
	case "scan":
		path := "."
		if len(os.Args) >= 3 {
			path = os.Args[2]
		}

		result, err := scanner.Scan(path)
		if err != nil {
			fmt.Println("Scan failed:", err)
			os.Exit(1)
		}

		err = scanner.WriteManifest(result, "output")
		if err != nil {
			fmt.Println("Write manifest failed:", err)
			os.Exit(1)
		}

		fmt.Println("Scan completed.")

	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Usage: repo2ai scan <path>")
	}
}
