package container

import (
	"testing"
)

func TestContainerAuth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		username  string
		password  string
		expectErr bool
		expectMsg string
	}{
		{
			name:      "valid credentials",
			username:  "admin",
			password:  "admin",
			expectErr: false,
			expectMsg: "success",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Auth(tc.username, tc.password)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
