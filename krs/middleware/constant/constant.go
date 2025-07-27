package constant

const (
	HeaderAuthorizationKey   = "Authorization"
	AuthorizationValueBearer = "Bearer"
	ErrMissingToken          = "missing token"
	ErrInvalidToken          = "invalid token"
	ErrMessageUnauthorized   = "Unauthorized"

	// default context key for jwt
	DefaultJwtContextKey = "user_id"
	// default context key for token
	DefaultTokenContextKey = "token"
	// default context key for file
	DefaultFileContextKey = "file"
)
