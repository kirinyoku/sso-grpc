// Package jwt provides JWT (JSON Web Token) functionality for authentication and authorization.
package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kirinyoku/sso-grpc/internal/domain/models"
)

// NewToken generates a JWT token for the specified user and application.
//
// Parameters:
//   - user: user to generate token for
//   - app: application to generate token for
//   - duration: duration for which the token is valid
//
// Returns:
//   - string: JWT token for authenticated sessions
//   - error: nil on success, or an error if token generation fails
func NewToken(user *models.User, app *models.App, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	calims := token.Claims.(jwt.MapClaims)

	calims["user_id"] = user.ID
	calims["app_id"] = app.ID
	calims["email"] = user.Email
	calims["exp"] = time.Now().Add(duration).Unix()

	return token.SignedString([]byte(app.Secret))
}
