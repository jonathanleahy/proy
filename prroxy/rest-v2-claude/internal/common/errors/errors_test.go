package errors_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/jonathanleahy/prroxy/rest-v2/internal/common/errors"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		appErr  *apperrors.AppError
		want    string
	}{
		{
			name: "error with underlying error",
			appErr: &apperrors.AppError{
				Code:    "TEST_ERROR",
				Message: "Test message",
				Err:     errors.New("underlying error"),
				Status:  500,
			},
			want: "Test message: underlying error",
		},
		{
			name: "error without underlying error",
			appErr: &apperrors.AppError{
				Code:    "TEST_ERROR",
				Message: "Test message",
				Status:  400,
			},
			want: "Test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appErr.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	appErr := &apperrors.AppError{
		Code:    "TEST_ERROR",
		Message: "Test message",
		Err:     underlyingErr,
		Status:  500,
	}

	unwrapped := errors.Unwrap(appErr)
	assert.Equal(t, underlyingErr, unwrapped)
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *apperrors.AppError
		wantCode   string
		wantStatus int
	}{
		{
			name:       "ErrNotFound",
			err:        apperrors.ErrNotFound,
			wantCode:   "NOT_FOUND",
			wantStatus: 404,
		},
		{
			name:       "ErrBadRequest",
			err:        apperrors.ErrBadRequest,
			wantCode:   "BAD_REQUEST",
			wantStatus: 400,
		},
		{
			name:       "ErrInternal",
			err:        apperrors.ErrInternal,
			wantCode:   "INTERNAL_ERROR",
			wantStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, tt.wantStatus, tt.err.Status)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name           string
		baseErr        *apperrors.AppError
		underlyingErr  error
		wantMessage    string
		wantCode       string
		wantStatus     int
	}{
		{
			name:          "wrap not found error",
			baseErr:       apperrors.ErrNotFound,
			underlyingErr: errors.New("user not found in database"),
			wantMessage:   "Resource not found",
			wantCode:      "NOT_FOUND",
			wantStatus:    404,
		},
		{
			name:          "wrap internal error",
			baseErr:       apperrors.ErrInternal,
			underlyingErr: errors.New("database connection failed"),
			wantMessage:   "Internal server error",
			wantCode:      "INTERNAL_ERROR",
			wantStatus:    500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := apperrors.Wrap(tt.baseErr, tt.underlyingErr)

			assert.Equal(t, tt.wantCode, wrapped.Code)
			assert.Equal(t, tt.wantStatus, wrapped.Status)
			assert.Equal(t, tt.wantMessage, wrapped.Message)
			assert.Equal(t, tt.underlyingErr, wrapped.Err)

			// Verify error message includes underlying error
			assert.Contains(t, wrapped.Error(), tt.underlyingErr.Error())
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		message    string
		status     int
	}{
		{
			name:    "custom error",
			code:    "CUSTOM_ERROR",
			message: "This is a custom error",
			status:  422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := apperrors.New(tt.code, tt.message, tt.status)

			assert.Equal(t, tt.code, err.Code)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.status, err.Status)
			assert.Nil(t, err.Err)
		})
	}
}
