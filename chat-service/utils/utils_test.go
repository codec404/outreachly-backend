package utils_test

import (
	"testing"

	"github.com/codec404/chat-service/utils"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		email string
		want  bool
	}{
		{
			name:  "valid email",
			email: "12test@example.com",
			want:  true,
		},
		{
			name:  "invalid email",
			email: "invalid-email",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsValidEmail(tt.email)
			if got != tt.want {
				t.Errorf("IsValidEmail() = %v, want %v", got, tt.want)
			} else {
				t.Logf("IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
