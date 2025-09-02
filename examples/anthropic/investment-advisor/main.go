package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	fmt.Println("=== AI Stock Market Analyst ===\n")

	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("Please set ANTHROPIC_API_KEY environment variable")
	}

	// Create agents for each component
	macroAgent, err := createMacroAnalysisAgent(apiKey)
	if err != nil {
		log.Fatal("Failed to create macro analysis agent:", err)
	}

	fundamentalAgent, err := createFundamentalAnalysisAgent(apiKey)
	if err != nil {
		log.Fatal("Failed to create fundamental analysis agent:", err)
	}

	technicalAgent, err := createTechnicalAnalysisAgent(apiKey)
	if err != nil {
		log.Fatal("Failed to create technical analysis agent:", err)
	}

	portfolioAgent, err := createPortfolioManagementAgent(apiKey)
	if err != nil {
		log.Fatal("Failed to create portfolio management agent:", err)
	}

	orchestrator, err := createInvestmentOrchestrator(apiKey)
	if err != nil {
		log.Fatal("Failed to create investment orchestrator:", err)
	}

	// Example investment analysis workflow
	fmt.Println("🔍 Running comprehensive investment analysis...\n")

	// Example: Analyze financial sector stocks
	symbols := []string{"JPM", "BAC", "WFC", "C", "GS"}

	for _, symbol := range symbols {
		fmt.Printf("📊 Analyzing %s...\n", symbol)

		// Macro analysis
		macroResp, err := macroAgent.Chat(fmt.Sprintf("Analyze the macro environment for financial sector stock %s. Consider interest rate trends, regulatory environment, and economic indicators.", symbol))
		if err != nil {
			log.Printf("Macro analysis failed for %s: %v", symbol, err)
			continue
		}

		// Fundamental analysis
		fundamentalResp, err := fundamentalAgent.Chat(fmt.Sprintf("Perform fundamental analysis on %s. Evaluate valuation metrics, financial health, and growth prospects.", symbol))
		if err != nil {
			log.Printf("Fundamental analysis failed for %s: %v", symbol, err)
			continue
		}

		// Technical analysis
		technicalResp, err := technicalAgent.Chat(fmt.Sprintf("Analyze technical indicators for %s. Look for oversold conditions and optimal entry points.", symbol))
		if err != nil {
			log.Printf("Technical analysis failed for %s: %v", symbol, err)
			continue
		}

		// Integration and final recommendation
		analysisPrompt := fmt.Sprintf(`
Based on the following analysis for %s, provide a final investment recommendation:

MACRO ANALYSIS:
%s

FUNDAMENTAL ANALYSIS:
%s

TECHNICAL ANALYSIS:
%s

Provide a structured recommendation following the Triangula Capital strategy:
1. Overall score (0-100)
2. Recommendation (BUY/HOLD/SELL)
3. Rationale
4. Risk assessment
5. Position sizing suggestion (if BUY)
`, symbol, macroResp.Text, fundamentalResp.Text, technicalResp.Text)

		finalResp, err := orchestrator.Chat(analysisPrompt)
		if err != nil {
			log.Printf("Final analysis failed for %s: %v", symbol, err)
			continue
		}

		fmt.Printf("💡 %s Analysis Complete:\n%s\n\n", symbol, finalResp.Text)

		// Portfolio management
		if strings.Contains(strings.ToUpper(finalResp.Text), "BUY") {
			portfolioResp, err := portfolioAgent.Chat(fmt.Sprintf("Consider adding %s to the portfolio based on this analysis: %s", symbol, finalResp.Text))
			if err != nil {
				log.Printf("Portfolio analysis failed for %s: %v", symbol, err)
			} else {
				fmt.Printf("📈 Portfolio Recommendation for %s:\n%s\n\n", symbol, portfolioResp.Text)
			}
		}
	}

	// Portfolio review
	fmt.Println("📋 Current Portfolio Review...")
	portfolioReview, err := portfolioAgent.Chat("Provide a comprehensive portfolio review including current positions, risk metrics, and rebalancing recommendations.")
	if err != nil {
		log.Printf("Portfolio review failed: %v", err)
	} else {
		fmt.Printf("Portfolio Status:\n%s\n", portfolioReview.Text)
	}
}
