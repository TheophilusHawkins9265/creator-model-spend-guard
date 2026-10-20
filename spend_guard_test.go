package main

import "testing"

func TestAllowProcessing(t *testing.T) {
	tests := []struct {
		name          string
		usage         Usage
		cap, estimate float64
		wantErr       bool
	}{
		{"under cap", Usage{SpendUSD: 4}, 5, 1, false},
		{"at cap", Usage{SpendUSD: 4}, 5, 1, false},
		{"over cap", Usage{SpendUSD: 4.01}, 5, 1, true},
		{"negative estimate", Usage{SpendUSD: 1}, 5, -1, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if (AllowProcessing(tc.usage, tc.cap, tc.estimate) != nil) != tc.wantErr {
				t.Fatalf("unexpected decision")
			}
		})
	}
}
