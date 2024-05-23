package apperrors

import (
	"fmt"
	"time"
)

var (
	_ error = (*AppError)(nil)
	_ error = (*AppErrorTransient)(nil)
	_ error = (*AppErrorInvalid)(nil)
	_ error = (*AppErrorNotFound)(nil)
	_ error = (*AppErrorInternal)(nil)
)

type (
	AppError struct {
		msg string
	}
	AppErrorTransient struct {
		AppError
		RetryAfter time.Duration
	}
	AppErrorInvalid struct {
		AppError
	}
	AppErrorNotFound struct {
		AppError
	}
	AppErrorInternal struct {
		AppError
	}
)

func NewTransient(
	msg string,
	retryAfter time.Duration,
) *AppErrorTransient {
	return &AppErrorTransient{
		AppError: AppError{
			msg: msg,
		},
		RetryAfter: retryAfter,
	}
}

func NewInvalid(msg string) *AppErrorInvalid {
	return &AppErrorInvalid{
		AppError: AppError{msg: msg},
	}
}

func NewInternal(msg string) *AppErrorInternal {
	return &AppErrorInternal{
		AppError{msg: msg},
	}
}

func NewNotFound(msg string) *AppErrorNotFound {
	return &AppErrorNotFound{
		AppError: AppError{msg: msg},
	}
}

func (e AppError) Error() string {
	return e.msg
}

func (e AppErrorTransient) Error() string {
	return fmt.Sprintf("%s, retry after %s", e.msg, e.RetryAfter.String())
}
