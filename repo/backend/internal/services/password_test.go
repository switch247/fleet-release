package services

import "testing"

func TestValidatePasswordComplexityPermutations(t *testing.T) {
	cases := []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "valid baseline", password: "ValidPassword123!", valid: true},
		{name: "too short", password: "V1!short", valid: false},
		{name: "missing uppercase", password: "validpassword123!", valid: false},
		{name: "missing lowercase", password: "VALIDPASSWORD123!", valid: false},
		{name: "missing digit", password: "ValidPassword!!!", valid: false},
		{name: "missing symbol", password: "ValidPassword1234", valid: false},
		{name: "unicode symbol still valid", password: "ValidPass123€A", valid: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePasswordComplexity(tc.password)
			if tc.valid && err != nil {
				t.Fatalf("expected valid password, got error: %v", err)
			}
			if !tc.valid && err == nil {
				t.Fatalf("expected invalid password")
			}
		})
	}
}
