package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aktagon/llmkit/anthropic"
	"github.com/aktagon/llmkit/anthropic/agents"
)

// mockWeatherAPI simulates a weather API call
func mockWeatherAPI(location string, unit string) (string, error) {
	// Normalize location for consistent responses
	location = strings.ToLower(strings.TrimSpace(location))

	// Mock weather data for demo purposes
	weatherData := map[string]map[string]string{
		"san francisco": {
			"celsius":    "18°C, partly cloudy with light fog",
			"fahrenheit": "64°F, partly cloudy with light fog",
		},
		"new york": {
			"celsius":    "22°C, sunny with clear skies",
			"fahrenheit": "72°F, sunny with clear skies",
		},
		"london": {
			"celsius":    "15°C, overcast with light rain",
			"fahrenheit": "59°F, overcast with light rain",
		},
		"tokyo": {
			"celsius":    "25°C, humid with scattered clouds",
			"fahrenheit": "77°F, humid with scattered clouds",
		},
		"helsinki": {
			"celsius":    "-2°C, light snow with overcast skies",
			"fahrenheit": "28°F, light snow with overcast skies",
		},
	}

	// Normalize unit
	if unit == "" {
		unit = "fahrenheit"
	}
	unit = strings.ToLower(unit)

	// Check for location match (flexible matching)
	for city, data := range weatherData {
		if strings.Contains(location, city) || strings.Contains(city, location) {
			if weather, ok := data[unit]; ok {
				return weather, nil
			}
			// Default to fahrenheit if unit not found
			return data["fahrenheit"], nil
		}
	}

	// Default response for unknown locations
	if unit == "celsius" {
		return "20°C, conditions unknown for this location", nil
	}
	return "68°F, conditions unknown for this location", nil
}

// createWeatherTool creates a weather tool for the chat agent
func createWeatherTool() anthropic.Tool {
	return anthropic.Tool{
		Name:        "get_weather",
		Description: "Get the current weather in a given location",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"location": map[string]interface{}{
					"type":        "string",
					"description": "The city and state, e.g. San Francisco, CA",
				},
				"unit": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"celsius", "fahrenheit"},
					"description": "The unit of temperature, either 'celsius' or 'fahrenheit'",
				},
			},
			"required": []string{"location"},
		},
		Handler: func(input map[string]interface{}) (string, error) {
			// Extract location (required)
			location, ok := input["location"].(string)
			if !ok || location == "" {
				return "", fmt.Errorf("location is required and must be a string")
			}

			// Extract unit (optional, defaults to fahrenheit)
			unit := "fahrenheit"
			if unitInput, exists := input["unit"]; exists {
				if unitStr, ok := unitInput.(string); ok {
					unit = unitStr
				}
			}

			// Call weather API (mocked)
			weather, err := mockWeatherAPI(location, unit)
			if err != nil {
				return "", fmt.Errorf("failed to get weather for %s: %w", location, err)
			}

			return fmt.Sprintf("The weather in %s is %s", location, weather), nil
		},
	}
}

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	// Create chat agent
	agent, err := agents.New(apiKey)
	if err != nil {
		log.Fatalf("Failed to create chat agent: %v", err)
	}

	// Register weather tool
	weatherTool := createWeatherTool()
	err = agent.RegisterTool(weatherTool)
	if err != nil {
		log.Fatalf("Failed to register weather tool: %v", err)
	}

	fmt.Println("=== Weather Chat Agent Demo ===")
	fmt.Println("Ask me about the weather in different cities!")
	fmt.Println("Try: 'What's the weather in San Francisco?'")
	fmt.Println("Type 'exit' to quit.")

	// Create scanner for reading full lines
	scanner := bufio.NewScanner(os.Stdin)

	// Interactive chat loop
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "" {
			continue
		}

		// Chat with agent (tools execute automatically)
		response, err := agent.Chat(input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Agent: %s\n\n", response.Text)
	}
}
