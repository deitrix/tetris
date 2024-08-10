package main

import (
	"slices"
	"testing"
)

func TestPiece_TrimSpace(t *testing.T) {
	tests := []struct {
		input  []int
		expect []int
	}{
		{
			input: []int{
				0, 1, 0,
				1, 1, 1,
				0, 0, 0,
			},
			expect: []int{
				0, 1, 0,
				1, 1, 1,
			},
		},
		{
			input: []int{
				0, 0, 0,
				0, 1, 0,
				0, 0, 0,
			},
			expect: []int{
				1,
			},
		},
		{
			input: []int{
				1, 0, 0,
				1, 1, 1,
				0, 0, 0,
			},
			expect: []int{
				1, 0, 0,
				1, 1, 1,
			},
		},
	}
	for _, test := range tests {
		p := Piece{
			Mask: test.input,
		}
		got := p.TrimSpace()
		if !slices.Equal(got.Mask, test.expect) {
			t.Errorf("TrimSpace(%v): got: %v, want: %v", p, got, test.expect)
		}
	}
}
