package main

import (
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

	matchStats, err := parser.ParseDemo(demoPath)
	if err != nil {
		log.Fatalf("Parsing error: %v", err)
	}

	jsonData, err := json.MarshalIndent(matchStats, "", "  ")
	fmt.Println(string(jsonData))
}
