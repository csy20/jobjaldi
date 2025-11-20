package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/csy/jobagent/jobagent"
)

// Import the module in a way that keeps it simple

func main() {
	// Test single provider scrape
	fmt.Println("=== Testing Single Provider Scrape ===")
	result, err := jobagent.ScrapeProvider("greenhouse", "stripe")
	if err != nil {
		log.Printf("Error scraping Stripe: %v\n", err)
	} else {
		fmt.Printf("Stripe jobs JSON length: %d bytes\n", len(result))
		var jobs []map[string]interface{}
		if err := json.Unmarshal([]byte(result), &jobs); err != nil {
			log.Printf("Error unmarshaling: %v\n", err)
		} else {
			fmt.Printf("Found %d jobs from Stripe\n", len(jobs))
			if len(jobs) > 0 {
				fmt.Printf("First job: %+v\n", jobs[0])
			}
		}
	}

	fmt.Println("\n=== Testing ScrapeMany ===")
	config := `{"targets":[
		{"provider":"greenhouse","company":"stripe"},
		{"provider":"greenhouse","company":"airbnb"}
	]}`

	result, err = jobagent.ScrapeMany(config)
	if err != nil {
		log.Printf("Error with ScrapeMany: %v\n", err)
	} else {
		fmt.Printf("ScrapeMany result length: %d bytes\n", len(result))
		var jobs []map[string]interface{}
		if err := json.Unmarshal([]byte(result), &jobs); err != nil {
			log.Printf("Error unmarshaling: %v\n", err)
		} else {
			fmt.Printf("Found %d total jobs\n", len(jobs))
		}
	}
}
