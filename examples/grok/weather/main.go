package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aktagon/llmkit/grok"
	"github.com/aktagon/llmkit/grok/types"
)

// WeatherQuery represents a parsed weather query
type WeatherQuery struct {
	Location string `json:"location"`
	Unit     string `json:"unit"`
}

// mockWeatherAPI simulates a weather API call
func mockWeatherAPI(location string, unit string) string {
	location = strings.ToLower(strings.TrimSpace(location))

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
	}

	if unit == "" {
		unit = "fahrenheit"
	}
	unit = strings.ToLower(unit)

	for city, data := range weatherData {
		if strings.Contains(location, city) || strings.Contains(city, location) {
			if weather, ok := data[unit]; ok {
				return weather
			}
			return data["fahrenheit"]
		}
	}

	if unit == "celsius" {
		return "20°C, conditions unknown"
	}
	return "68°F, conditions unknown"
}

// extractWeatherQuery uses Grok to parse natural language weather queries
func extractWeatherQuery(userInput, apiKey string) (*WeatherQuery, error) {
	schema := `{
		"name": "weather_query",
		"description": "Extract location and temperature unit from weather query",
		"strict": true,
		"schema": {
			"type": "object",
			"properties": {
				"location": {
					"type": "string",
					"description": "The city or location name"
				},
				"unit": {
					"type": "string",
					"enum": ["celsius", "fahrenheit"],
					"description": "Temperature unit preference"
				}
			},
			"required": ["location"],
			"additionalProperties": false
		}
	}`

	systemPrompt := "Extract the location and temperature unit from the user's weather query."
	response, err := grok.Prompt(systemPrompt, userInput, schema, apiKey)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from Grok")
	}

	var query WeatherQuery
	err = json.Unmarshal([]byte(response.Choices[0].Message.Content), &query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse weather query: %w", err)
	}

	return &query, nil
}

// formatWeatherResponse uses Grok to create a natural response
func formatWeatherResponse(location, weather, apiKey string) (string, error) {
	systemPrompt := "You are a helpful weather assistant. Create a natural, friendly response about the weather."
	userPrompt := fmt.Sprintf("The weather in %s is %s. Provide a brief, friendly response.", location, weather)

	settings := types.RequestSettings{
		MaxTokens:   100,
		Temperature: 0.7,
	}

	response, err := grok.PromptWithSettings(systemPrompt, userPrompt, "", apiKey, settings)
	if err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from Grok")
	}

	return response.Choices[0].Message.Content, nil
}

func main() {
	apiKey := os.Getenv("XAI_API_KEY")
	if apiKey == "" {
		log.Fatal("XAI_API_KEY environment variable is required")
	}

	fmt.Println("=== Grok Weather Assistant ===")
	fmt.Println("Ask me about the weather in different cities!")
	fmt.Println("Try: 'What's the weather in San Francisco?'")
	fmt.Println("     'How's the weather in Tokyo in celsius?'")
	fmt.Println("Type 'exit' to quit.\n")

	scanner := bufio.NewScanner(os.Stdin)

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

		// Step 1: Extract location and unit from query
		query, err := extractWeatherQuery(input, apiKey)
		if err != nil {
			fmt.Printf("Error parsing query: %v\n\n", err)
			continue
		}

		// Step 2: Get weather data (mocked)
		weather := mockWeatherAPI(query.Location, query.Unit)

		// Step 3: Format natural language response
		response, err := formatWeatherResponse(query.Location, weather, apiKey)
		if err != nil {
			fmt.Printf("Error formatting response: %v\n\n", err)
			continue
		}

		fmt.Printf("\nGrok: %s\n\n", response)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}