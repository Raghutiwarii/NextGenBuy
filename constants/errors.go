package constants

const (
	ErrorInvalidRequestPayload = 1001
	ErrorDatabaseQueryFailed   = 1002
	ErrorDatabaseUpdateFailed  = 1003
	ErrorDatabaseCreateFailed  = 1004
	ErrorInvalidOTP            = 1005
	ErrorUserAlreadyExists     = 1006
	ErrorHashingFailed         = 1007
	ErrorUserNotFound          = 1008
	ErrorInvalidCredentials    = 1009
	ErrorTokenGenerationFailed = 1010
	ErrorUnauthorized          = 1011
)

func ErrorText(code int) string {
	switch code {
	case ErrorInvalidRequestPayload:
		return "Invalid request payload"
	case ErrorDatabaseQueryFailed:
		return "Failed to query database"
	case ErrorDatabaseUpdateFailed:
		return "Failed to update database"
	case ErrorDatabaseCreateFailed:
		return "Failed to create record in database"
	case ErrorInvalidOTP:
		return "Invalid OTP"
	case ErrorUserAlreadyExists:
		return "User already exists"
	case ErrorHashingFailed:
		return "Error while parsing password"
	case ErrorUserNotFound:
		return "User not found"
	case ErrorInvalidCredentials:
		return "Invalid credentials"
	case ErrorTokenGenerationFailed:
		return "Error while generating token"
	case ErrorUnauthorized:
		return "Unauthorized"
	default:
		return "Unknown error"
	}
}
