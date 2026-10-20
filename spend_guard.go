package main

import "fmt"

type Usage struct {
	SpendUSD float64
}

// AllowProcessing makes the spend decision from the latest account usage.
func AllowProcessing(usage Usage, hardCapUSD, estimateUSD float64) error {
	if estimateUSD < 0 {
		return fmt.Errorf("estimate must be non-negative")
	}
	if usage.SpendUSD+estimateUSD > hardCapUSD {
		return fmt.Errorf("hard cap would be exceeded")
	}
	return nil
}
