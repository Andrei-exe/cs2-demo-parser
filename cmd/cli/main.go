package main

import (
	"encoding/json"
	"fmt"
	"log"

	"cs2-parser/internal/parser"
)

func main() {
	demoPath := "demos/test_demo.dem"

	matchStats, err := parser.ParseDemo(demoPath)
	if err != nil {
		log.Fatalf("Parsing error: %v", err)
	}

	jsonData, err := json.MarshalIndent(matchStats, "", "  ")
	fmt.Println(string(jsonData))
}
