package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"cs2-parser/internal/parser"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Command syntax: go run ./cmd/cli/main.go <DEMO_FILE_PATH>")
	}
	demoPath := os.Args[1]

	matchStats, err := parser.ParseDemo(context.Background(), demoPath, func(f float64) {
		fmt.Fprintf(os.Stderr, "\rParsing... %3.0f%%", f*100)
	})
	fmt.Fprintln(os.Stderr)

	if err != nil {
		log.Fatalf("Parsing error: %v", err)
	}

	jsonData, _ := json.MarshalIndent(matchStats, "", "  ")
	fmt.Println(string(jsonData))
}
