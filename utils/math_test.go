package utils

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		// 正常情况
		{"1 + 2 = 3", 1, 2, 3},
		{"10 + 20 = 30", 10, 20, 30},
		{"100 + 200 = 300", 100, 200, 300},

		// 负数情况
		{"-1 + -2 = -3", -1, -2, -3},
		{"-5 + 3 = -2", -5, 3, -2},
		{"5 + -3 = 2", 5, -3, 2},

		// 零的情况
		{"0 + 0 = 0", 0, 0, 0},
		{"0 + 5 = 5", 0, 5, 5},
		{"5 + 0 = 5", 5, 0, 5},

		// 边界情况
		{"max int + 0", 2147483647, 0, 2147483647},
		{"min int + 0", -2147483648, 0, -2147483648},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Add(tt.a, tt.b)
			if got != tt.expected {
				t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, got, tt.expected)
			}
		})
	}
}
