package main

import (
	"github.com/aktagon/llmkit/anthropic/agents"
)

// createMacroAnalysisAgent creates an agent specialized in global macro analysis
func createMacroAnalysisAgent(apiKey string) (*agents.ChatAgent, error) {
	agent, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryPersistence("./data/macro_memory.json"),
	)
	if err != nil {
		return nil, err
	}

	// Register market data tool
	if err := registerMarketDataTool(agent); err != nil {
		return nil, err
	}

	// Register economic indicators tool
	if err := registerEconomicIndicatorsTool(agent); err != nil {
		return nil, err
	}

	// Set specialized system prompt
	agent.Remember("role", "Global Macro Analysis Specialist")
	agent.Remember("focus", "Identify distressed countries/sectors and macro opportunities")
	agent.Remember("expertise", "Economic indicators, central bank policies, geopolitical analysis")

	return agent, nil
}

// createFundamentalAnalysisAgent creates an agent specialized in fundamental analysis
func createFundamentalAnalysisAgent(apiKey string) (*agents.ChatAgent, error) {
	agent, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryPersistence("./data/fundamental_memory.json"),
	)
	if err != nil {
		return nil, err
	}

	// Register fundamental analysis tools
	if err := registerFundamentalScreeningTool(agent); err != nil {
		return nil, err
	}

	if err := registerValuationTool(agent); err != nil {
		return nil, err
	}

	agent.Remember("role", "Fundamental Analysis Specialist")
	agent.Remember("focus", "Computer-assisted screening for undervalued stocks")
	agent.Remember("expertise", "Financial statement analysis, valuation metrics, quality assessment")

	return agent, nil
}

// createTechnicalAnalysisAgent creates an agent specialized in technical analysis
func createTechnicalAnalysisAgent(apiKey string) (*agents.ChatAgent, error) {
	agent, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryPersistence("./data/technical_memory.json"),
	)
	if err != nil {
		return nil, err
	}

	// Register technical analysis tools
	if err := registerTechnicalIndicatorsTool(agent); err != nil {
		return nil, err
	}

	if err := registerChartPatternTool(agent); err != nil {
		return nil, err
	}

	agent.Remember("role", "Technical Analysis Specialist")
	agent.Remember("focus", "Identify oversold entry timing and optimal entry points")
	agent.Remember("expertise", "Technical indicators, chart patterns, momentum analysis")

	return agent, nil
}

// createPortfolioManagementAgent creates an agent specialized in portfolio management
func createPortfolioManagementAgent(apiKey string) (*agents.ChatAgent, error) {
	agent, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryPersistence("./data/portfolio_memory.json"),
	)
	if err != nil {
		return nil, err
	}

	// Register portfolio management tools
	if err := registerPortfolioTrackingTool(agent); err != nil {
		return nil, err
	}

	if err := registerRiskManagementTool(agent); err != nil {
		return nil, err
	}

	agent.Remember("role", "Portfolio Management Specialist")
	agent.Remember("focus", "Maintain concentrated 10-15 position portfolio")
	agent.Remember("constraints", "Max 15 positions, 2-6 month holding periods, accept 15-25% drawdowns")
	agent.Remember("objective", "MSCI World outperformance over 5+ years")

	return agent, nil
}

// createInvestmentOrchestrator creates the main coordination agent
func createInvestmentOrchestrator(apiKey string) (*agents.ChatAgent, error) {
	agent, err := agents.New(apiKey,
		agents.WithMemoryContext(),
		agents.WithMemoryPersistence("./data/orchestrator_memory.json"),
	)
	if err != nil {
		return nil, err
	}

	agent.Remember("role", "Investment Decision Orchestrator")
	agent.Remember("strategy", "Three-pillar value identification")
	agent.Remember("approach", "Combine macro + fundamental + technical analysis")
	agent.Remember("requirements", "Require alignment from at least 2/3 core analysis pillars")

	return agent, nil
}
