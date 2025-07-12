package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aktagon/llmkit/openai"
)

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Check if file exists
	filePath := "berkshire-10-K-pages-1.pdf"
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", filePath)
	}

	// Upload the PDF file
	file, err := openai.UploadFile(filePath, "assistants", apiKey)
	if err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}

	// Extract the risk section
	response, err := openai.Prompt(
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