package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/anthropic"
)

func main() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Upload the PDF file
	file, err := anthropic.UploadFile("berkshire-10-K-pages-1.pdf", apiKey)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	// Extract the risk section
	response, err := anthropic.Prompt(
		"You are a financial analyst expert at extracting information from SEC filings.",
		"Extract the 1A risk section.",
		"",
		apiKey,
		*file,
	)
	if err != nil {
		log.Fatalf("Failed to analyze document: %v", err)
	}

	fmt.Println(response)
}
