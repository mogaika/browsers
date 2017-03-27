package main

import (
	"log"

	"github.com/mogaika/browsers"

	// Use default browsers set
	// You can manually add or remove browser interfaces
	_ "github.com/mogaika/browsers/default"
)

/*
	This example print all your saved passwords
*/
func main() {
	passwds, errs := browsers.SavedPasswords()
	for _, err := range errs {
		log.Printf("Error: %v", err)
	}
	for _, p := range passwds {
		log.Println(p)
	}
}
