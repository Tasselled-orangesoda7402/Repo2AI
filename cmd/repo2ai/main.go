package main

import (
	"fmt"
	"os"

	"github.com/fichil/Repo2AI/internal/packer"
	"github.com/fichil/Repo2AI/internal/scanner"
	"strconv"
	"strings"
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

		for i := 3; i < len(os.Args); i++ {
			arg := strings.ToLower(os.Args[i])

			if strings.HasPrefix(arg, "--max-size=") {
				sizeText := strings.TrimPrefix(arg, "--max-size=")

				size, err := parseSize(sizeText)
				if err == nil {
					packer.SetMaxPackSize(size)
				}
			}
		}

		result, err := scanner.Scan(path)
		if err != nil {
			fmt.Println("Scan failed:", err)
			os.Exit(1)
		}

		err = packer.CleanOutput()
		if err != nil {
			fmt.Println("Clean output failed:", err)
			os.Exit(1)
		}

		err = scanner.WriteManifest(result, "output")
		if err != nil {
			fmt.Println("Write manifest failed:", err)
			os.Exit(1)
		}

		err = packer.Generate(result)
		if err != nil {
			fmt.Println("Generate context pack failed:", err)
			os.Exit(1)
		}

		fmt.Println("Scan completed.")

	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Usage: repo2ai scan <path>")
	}
}

func parseSize(text string) (int, error) {
	text = strings.ToLower(strings.TrimSpace(text))

	if strings.HasSuffix(text, "mb") {
		num := strings.TrimSuffix(text, "mb")
		n, err := strconv.Atoi(num)
		if err != nil {
			return 0, err
		}
		return n * 1024 * 1024, nil
	}

	if strings.HasSuffix(text, "kb") {
		num := strings.TrimSuffix(text, "kb")
		n, err := strconv.Atoi(num)
		if err != nil {
			return 0, err
		}
		return n * 1024, nil
	}

	return strconv.Atoi(text)
}
