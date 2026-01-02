package src

import (
	"fmt"
	"testing"
)

func TestPaginate(t *testing.T) {
	tests := []struct {
		name          string
		totalItems    int
		currentPage   int
		pageSize      int
		expectedStart int
		expectedEnd   int
		expectedBtns  []string
	}{
		{
			name:          "Page 1 of 1 (Total 5, Size 7)",
			totalItems:    5,
			currentPage:   1,
			pageSize:      7,
			expectedStart: 0,
			expectedEnd:   5,
			expectedBtns:  []string{"· 1 ·"},
		},
		{
			name:          "Page 1 of 2 (Total 10, Size 7)",
			totalItems:    10,
			currentPage:   1,
			pageSize:      7,
			expectedStart: 0,
			expectedEnd:   7,
			expectedBtns:  []string{"· 1 ·", "2", "Next >"},
		},
		{
			name:          "Page 2 of 2 (Total 10, Size 7)",
			totalItems:    10,
			currentPage:   2,
			pageSize:      7,
			expectedStart: 7,
			expectedEnd:   10,
			expectedBtns:  []string{"< Prev", "1", "· 2 ·"},
		},
		{
			name:          "Page 1 of 3 (Total 20, Size 7)",
			totalItems:    20,
			currentPage:   1,
			pageSize:      7,
			expectedStart: 0,
			expectedEnd:   7,
			expectedBtns:  []string{"· 1 ·", "2", "3", "Next >"},
		},
		{
			name:          "Page 2 of 3 (Total 20, Size 7)",
			totalItems:    20,
			currentPage:   2,
			pageSize:      7,
			expectedStart: 7,
			expectedEnd:   14,
			expectedBtns:  []string{"< Prev", "1", "· 2 ·", "3", "Next >"},
		},
		{
			name:          "Page 3 of 3 (Total 20, Size 7)",
			totalItems:    20,
			currentPage:   3,
			pageSize:      7,
			expectedStart: 14,
			expectedEnd:   20,
			expectedBtns:  []string{"< Prev", "1", "2", "· 3 ·"},
		},
		{
			name:          "Page 5 of 10 (Total 70, Size 7)",
			totalItems:    70,
			currentPage:   5,
			pageSize:      7,
			expectedStart: 28,
			expectedEnd:   35,
			expectedBtns:  []string{"< Prev", "4", "· 5 ·", "6", "Next >"},
		},
		{
			name:          "Page 10 of 10 (Total 70, Size 7)",
			totalItems:    70,
			currentPage:   10,
			pageSize:      7,
			expectedStart: 63,
			expectedEnd:   70,
			expectedBtns:  []string{"< Prev", "8", "9", "· 10 ·"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, buttons := Paginate(tt.totalItems, tt.currentPage, tt.pageSize, "cb:")
			if start != tt.expectedStart {
				t.Errorf("start = %d, want %d", start, tt.expectedStart)
			}
			if end != tt.expectedEnd {
				t.Errorf("end = %d, want %d", end, tt.expectedEnd)
			}

			if len(buttons) != len(tt.expectedBtns) {
				t.Errorf("got %d buttons, want %d", len(buttons), len(tt.expectedBtns))
				for _, b := range buttons {
					fmt.Printf("Got button: %s\n", b.Text)
				}
				return
			}
			for i, btn := range buttons {
				if btn.Text != tt.expectedBtns[i] {
					t.Errorf("button[%d] text = %s, want %s", i, btn.Text, tt.expectedBtns[i])
				}
			}
		})
	}
}
