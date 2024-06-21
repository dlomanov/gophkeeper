package entities

import "errors"

type (
	SignUpUserRequest struct {
		Login    string
		Password string
	}
	SignInUserRequest struct {
		Login    string
		Password string
	}
)

func (r SignInUserRequest) Validate() (err error) {
	if r.Login == "" {
		err = errors.Join(err, ErrUserLoginInvalid)
	}
	if r.Password == "" {
		err = errors.Join(err, ErrUserPasswordInvalid)
	}
	return err
}

func (r SignUpUserRequest) Validate() (err error) {
	if r.Login == "" {
		err = errors.Join(err, ErrUserLoginInvalid)
	}
	if r.Password == "" {
		err = errors.Join(err, ErrUserPasswordInvalid)
	}
	return err
}
