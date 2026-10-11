package main

import (
	"encoding/json"
	"fmt"
	"log"

	"cs2-parser/internal/parser"
)

func main() {
	caleDemo := "demos/test_demo.dem"

	// 1. Apelăm parserul și prindem datele
	matchStats, err := parser.ParseDemo(caleDemo)
	if err != nil {
		log.Fatalf("Eroare la parsare, frățică: %v", err)
	}

	// 2. Transformăm structura Go într-un JSON formatat frumos (cu spații și indentare)
	jsonDate, err := json.MarshalIndent(matchStats, "", "  ")
	if err != nil {
		log.Fatalf("Nu am putut face JSON-ul: %v", err)
	}

	// 3. Afișăm în terminal rezultatul final
	fmt.Println("Gata șefule! Uite ce a ieșit:")
	fmt.Println(string(jsonDate))
}
