package usecases_test

import (
	"context"
	"errors"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/services/pass"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/services/token"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"testing"
	"time"
)

const (
	ActionRegister = "register"
	ActionLogin    = "login"
)

func TestUserUC(t *testing.T) {
	type args struct {
		action string
		creds  entities.Creds
	}
	type want struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "fail: empty login",
			args: args{
				action: ActionRegister,
				creds: entities.Creds{
					Login: "",
					Pass:  "1",
				},
			},
			want: want{err: entities.ErrUserCredsInvalid},
		},
		{
			name: "fail: empty password",
			args: args{
				action: ActionRegister,
				creds: entities.Creds{
					Login: "admin",
					Pass:  "",
				},
			},
			want: want{err: entities.ErrUserCredsInvalid},
		},
		{
			name: "fail empty creds",
			args: args{
				action: ActionRegister,
				creds: entities.Creds{
					Login: "",
					Pass:  "",
				},
			},
			want: want{err: entities.ErrUserCredsInvalid},
		},
		{
			name: "success: user registered",
			args: args{
				action: ActionRegister,
				creds: entities.Creds{
					Login: "admin",
					Pass:  "1",
				},
			},
			want: want{err: nil},
		},
		{
			name: "failed: user already registered",
			args: args{
				action: ActionRegister,
				creds: entities.Creds{
					Login: "admin",
					Pass:  "1",
				},
			},
			want: want{err: entities.ErrUserExists},
		},
		{
			name: "success: user registered",
			args: args{
				action: ActionRegister,
				creds: entities.Creds{
					Login: "admin2",
					Pass:  "1",
				},
			},
			want: want{err: nil},
		},
		{
			name: "fail: empty login",
			args: args{
				action: ActionLogin,
				creds: entities.Creds{
					Login: "admin",
					Pass:  "1",
				},
			},
			want: want{err: entities.ErrUserCredsInvalid},
		},
		{
			name: "fail: empty password",
			args: args{
				action: ActionLogin,
				creds: entities.Creds{
					Login: "admin",
					Pass:  "",
				},
			},
			want: want{err: entities.ErrUserCredsInvalid},
		},
		{
			name: "fail: empty creds",
			args: args{
				action: ActionLogin,
				creds: entities.Creds{
					Login: "admin",
					Pass:  "",
				},
			},
			want: want{err: entities.ErrUserCredsInvalid},
		},
		{
			name: "success: user logged in",
			args: args{
				action: ActionLogin,
				creds: entities.Creds{
					Login: "admin",
					Pass:  "1",
				},
			},
			want: want{err: nil},
		},
	}

	tokener := token.NewJWT([]byte("testsecret"), time.Minute)
	uc := usecases.NewUserUC(
		zaptest.NewLogger(t),
		NewMockUserRepo(),
		pass.NewHasher(0),
		tokener,
	)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotToken entities.Token
				err      error
			)

			switch tt.args.action {
			case ActionRegister:
				gotToken, err = uc.SignUp(ctx, tt.args.creds)
			case ActionLogin:
				gotToken, err = uc.SignIn(ctx, tt.args.creds)
			default:
				t.Fatalf("unknown action type: %s", tt.args.action)
			}

			if errors.Is(err, tt.want.err) {
				return
			}

			require.NoErrorf(t, err, "%s: unexpected error occured: '%v'", tt.args.action, err)
			require.NotEmptyf(t, gotToken, "%s: token should not be empty", tt.args.action)

			userID, err := tokener.GetUserID(gotToken)
			require.NoErrorf(t, err, "%s: error '%v' occured while extracting userID from token", tt.args.action, err)
			require.NotEmptyf(t, uuid.UUID(userID), "%s: userID should not be empty", tt.args.action)
		})
	}
}
