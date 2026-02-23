package errorType

import "net/http"

// AppError is the single error type used across the entire application.
// Every error that crosses a layer boundary should be an AppError.
//
// Code       → machine readable, used in logs and debugging ("EMAIL_ALREADY_REGISTERED")
// Message    → human readable, safe to send to the client
// HTTPStatus → maps directly to the HTTP response code
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"` // never serialised — only used internally
}

// Error implements the standard Go error interface.
// This means AppError can be used anywhere a regular error is expected.
func (e AppError) Error() string {
	return e.Message
}

// Is allows errors.Is() to work correctly with AppError.
// This is what makes `errors.Is(err, errorType.ErrUserNotFound)` work
// even when the error has been wrapped with fmt.Errorf("... %w", err).
func (e AppError) Is(target error) bool {
	t, ok := target.(AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// ── Auth Errors ───────────────────────────────────────────────────────────────

var (
	ErrEmailAlreadyRegistered = AppError{
		Code:       "EMAIL_ALREADY_REGISTERED",
		Message:    "email already registered",
		HTTPStatus: http.StatusConflict,
	}

	ErrInvalidCredentials = AppError{
		Code:       "INVALID_CREDENTIALS",
		Message:    "invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrEmailAlreadyVerified = AppError{
		Code:       "EMAIL_ALREADY_VERIFIED",
		Message:    "email is already verified",
		HTTPStatus: http.StatusConflict,
	}

	ErrInvalidOTP = AppError{
		Code:       "INVALID_OTP",
		Message:    "invalid OTP",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrOTPExpired = AppError{
		Code:       "OTP_EXPIRED",
		Message:    "OTP expired or not found — please request a new one",
		HTTPStatus: http.StatusBadRequest,
	}
	ErrEmailNotVerified = AppError{
		Code:       "EMAIL_NOT_VERIFIED",
		Message:    "email is not verified",
		HTTPStatus: http.StatusUnauthorized,
	}
	// authorization header format must be: Bearer <token>
	ErrTokenFormat = AppError{
		Code:       "INVALID_TOKEN",
		Message:    "invalid or expired token format",
		HTTPStatus: http.StatusUnauthorized,
	}
)

// ── User Errors ───────────────────────────────────────────────────────────────

var (
	ErrUserNotFound = AppError{
		Code:       "USER_NOT_FOUND",
		Message:    "user not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrFailedToCreateUser = AppError{
		Code:       "FAILED_TO_CREATE_USER",
		Message:    "failed to create user",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrFailedToUpdateUser = AppError{
		Code:       "FAILED_TO_UPDATE_USER",
		Message:    "failed to update user",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrFailedToDeleteUser = AppError{
		Code:       "FAILED_TO_DELETE_USER",
		Message:    "failed to delete user",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// ── Token Errors ──────────────────────────────────────────────────────────────

var (
	ErrInvalidToken = AppError{
		Code:       "INVALID_TOKEN",
		Message:    "invalid or expired token",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrFailedToGenerateToken = AppError{
		Code:       "FAILED_TO_GENERATE_TOKEN",
		Message:    "failed to generate token",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// ── General Errors ────────────────────────────────────────────────────────────

var (
	ErrInternalServer = AppError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "something went wrong",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrFailedToGenerateOTP = AppError{
		Code:       "FAILED_TO_GENERATE_OTP",
		Message:    "failed to generate OTP",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrFailedToStoreOTP = AppError{
		Code:       "FAILED_TO_STORE_OTP",
		Message:    "failed to store OTP",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrEmailServiceDown = AppError{
		Code:       "EMAIL_SERVICE_DOWN",
		Message:    "email service is down",
		HTTPStatus: http.StatusInternalServerError,
	}
	ErrFailedToEnqueueTask = AppError{
		Code:       "FAILED_TO_ENQUEUE_TASK",
		Message:    "failed to enqueue task",
		HTTPStatus: http.StatusInternalServerError,
	}
)
