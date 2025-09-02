# AI Stock Market Analyst

An AI-powered investment system implementing an investment strategy through a multi-agent architecture using the llmkit framework.

## Overview

This system implements a three-pillar investment approach:

1. **Global Macro Analysis** - Identifies distressed countries/sectors
2. **Fundamental Analysis** - Computer-assisted screening for undervalued stocks
3. **Technical Analysis** - Oversold entry timing optimization

The system maintains a concentrated portfolio of 10-15 high-conviction positions with 2-6 month holding periods, targeting MSCI World outperformance.

## Architecture

### Core Agents

- **Macro Analysis Agent** - Economic indicators, central bank policies, geopolitical analysis
- **Fundamental Analysis Agent** - Financial screening, valuation metrics, quality assessment
- **Technical Analysis Agent** - Technical indicators, chart patterns, momentum analysis
- **Portfolio Management Agent** - Position tracking, risk management, performance monitoring
- **Integration Orchestrator** - Coordinates all agents and synthesizes final recommendations

### Tools Available

Each agent has access to specialized tools:

#### Market Data Tools

- `get_market_data` - Real-time stock price and volume data
- `get_economic_indicators` - Economic indicators for macro analysis

#### Analysis Tools

- `screen_fundamentals` - Screen stocks based on fundamental criteria
- `calculate_valuation` - Comprehensive valuation metrics
- `analyze_technical_indicators` - Technical analysis for entry/exit timing
- `identify_chart_patterns` - Chart pattern recognition

#### Portfolio Tools

- `track_portfolio` - Portfolio position tracking and performance
- `assess_risk` - Risk assessment and alert generation

## Usage

### Prerequisites

1. Set your Anthropic API key:

   ```bash
   export ANTHROPIC_API_KEY="your-api-key-here"
   ```

2. Ensure you have Go 1.24+ installed

### Running the System

```bash
cd examples/anthropic/investment-advisor
go run main.go
```

### Example Output

The system will analyze financial sector stocks (JPM, BAC, WFC, C, GS) through each pillar:

```
=== AI Stock Market Analyst ===

🔍 Running comprehensive investment analysis...

📊 Analyzing JPM...
💡 JPM Analysis Complete:
Based on the comprehensive three-pillar analysis:
- Macro Score: 82/100 (Rising rates benefit financials)
- Fundamental Score: 85/100 (Strong ROE, reasonable valuation)
- Technical Score: 78/100 (Oversold, good entry point)
- Overall Score: 82/100
- Recommendation: BUY
- Position Size: 8-10% of portfolio
- Risk Level: Moderate

📈 Portfolio Recommendation for JPM:
Add JPM to portfolio with 8.5% allocation. Strong fundamental metrics
combined with favorable macro environment and oversold technical setup.
```

## Strategy Implementation

### Investment Criteria

- **Value Focus**: P/E < 15, P/B < 2.0, strong ROE
- **Quality Metrics**: Low debt-to-equity, high interest coverage
- **Technical Timing**: RSI < 30, support level confirmation
- **Macro Context**: Sector tailwinds, economic environment

### Risk Management

- Maximum 15 positions in portfolio
- Individual position limit: 15% of portfolio
- Sector concentration limit: 40%
- Drawdown alert at 20%, maximum tolerance 25%
- Stop-loss implementation for risk control

### Performance Targets

- Primary: Outperform MSCI World by 300+ basis points annually
- Sharpe Ratio: Target > 1.2
- Maximum Drawdown: < 25%
- Holding Period: 2-6 months average

## Data Sources

Current implementation uses mock data for demonstration. In production, integrate with:

- Financial data APIs (Alpha Vantage, Yahoo Finance, Bloomberg)
- Economic data feeds (FRED, ECB, Bank of Japan)
- News sentiment analysis services
- Real-time market data providers

## Memory and Persistence

Each agent maintains persistent memory for:

- Historical analysis patterns
- Market regime identification
- Performance tracking
- Risk assessment trends

Memory files are stored in `./data/` directory:

- `macro_memory.json` - Macro analysis patterns
- `fundamental_memory.json` - Valuation insights
- `technical_memory.json` - Technical patterns
- `portfolio_memory.json` - Portfolio state
- `orchestrator_memory.json` - Decision history

## Configuration

The system follows these strategy parameters:

- **Portfolio Size**: 10-15 positions maximum
- **Holding Period**: 2-6 months
- **Risk Tolerance**: 15-25% drawdowns accepted
- **Focus Sectors**: Currently emphasizing financials
- **Benchmark**: MSCI World Index

## Compliance and Risk

- All investment decisions require human oversight
- Comprehensive audit trail of all recommendations
- Risk monitoring and alert system
- Regular performance attribution analysis
- Stress testing under various market scenarios

## Extension Points

The modular architecture supports easy extension:

- Add new analysis tools
- Integrate additional data sources
- Implement custom risk models
- Create sector-specific agents
- Add backtesting capabilities

## Disclaimer

This system is for educational and research purposes. All investment decisions should be validated by qualified professionals. Past performance does not guarantee future results. The system accepts high volatility and significant drawdown risk as part of the strategy design.

---

Interested in AI-powered workflow automation for your company? Get started: https://aktagon.com | contact@aktagon.com
