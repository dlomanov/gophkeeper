package entities

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidate(t *testing.T) {
	type validator interface {
		Validate() error
	}
	tests := []struct {
		name     string
		r        validator
		wantErrs []error
	}{
		{
			name: "sign_up_user_request",
			r: SignUpUserRequest{
				Login:    "",
				Password: "",
			},
			wantErrs: []error{
				ErrUserLoginInvalid,
				ErrUserPasswordInvalid,
			},
		},
		{
			name: "sign_in_user_request",
			r: SignInUserRequest{
				Login:    "",
				Password: "",
			},
			wantErrs: []error{
				ErrUserLoginInvalid,
				ErrUserPasswordInvalid,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.Validate()
			for _, wantErr := range tt.wantErrs {
				require.ErrorIs(t, err, wantErr, "error mismatch")
			}
		})
	}
}
