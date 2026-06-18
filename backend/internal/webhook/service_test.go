package webhook

import "testing"

func TestShouldRetryStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		retry    bool
	}{
		{name: "retry on 500", status: 500, retry: true},
		{name: "do not retry on 524", status: 524, retry: false},
		{name: "retry on 429", status: 429, retry: true},
		{name: "retry on 408", status: 408, retry: true},
		{name: "do not retry on 400", status: 400, retry: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRetryStatus(tt.status)
			if got != tt.retry {
				t.Fatalf("shouldRetryStatus(%d) = %v, want %v", tt.status, got, tt.retry)
			}
		})
	}
}
