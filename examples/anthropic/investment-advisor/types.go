package main

import "time"

// Portfolio represents the current investment portfolio
type Portfolio struct {
	Positions   []Position `json:"positions"`
	Cash        float64    `json:"cash"`
	TotalValue  float64    `json:"total_value"`
	LastUpdated time.Time  `json:"last_updated"`
}

// Position represents a single stock position
type Position struct {
	Symbol    string    `json:"symbol"`
	Shares    float64   `json:"shares"`
	Price     float64   `json:"price"`
	Value     float64   `json:"value"`
	Weight    float64   `json:"weight"`
	EntryDate time.Time `json:"entry_date"`
}

// InvestmentAnalysis represents comprehensive analysis output
type InvestmentAnalysis struct {
	Symbol           string  `json:"symbol"`
	MacroScore       float64 `json:"macro_score"`
	FundamentalScore float64 `json:"fundamental_score"`
	TechnicalScore   float64 `json:"technical_score"`
	OverallScore     float64 `json:"overall_score"`
	Recommendation   string  `json:"recommendation"`
	Rationale        string  `json:"rationale"`
	RiskLevel        string  `json:"risk_level"`
}