package main

import (
	"encoding/json"
	"fmt"
	"log"

	"cs2-parser/internal/parser"
)

func main() {
	caleDemo := "demos/test_demo.dem"

	matchStats, err := parser.ParseDemo(caleDemo)
	if err != nil {
		log.Fatalf("Parsing error: %v", err)
	}

	jsonDate, err := json.MarshalIndent(matchStats, "", "  ")
	fmt.Println(string(jsonDate))
}
