package user

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"web-app/data"
	"web-app/httperror"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// jwtAccessMetadata stores information embedded in a JWT access token.
type jwtAccessMetadata struct {
	authUUID  string
	userID    uint
	createdAt time.Time
	expiresAt time.Time
}

// jwtRefreshMetadata stores information embedded in a JWT refresh token.
type jwtRefreshMetadata struct {
	authUUID  string
	userID    uint
	createdAt time.Time
	expiresAt time.Time
}

const (
	// authorizedFailedGeneric is returned when we are unable to authenticate
	// a user request.
	authorizationFailedGeneric = "request not authorized"
	// insufficientPermissionsGeneric is returned when a user does not have
	// required permissions to complete a request.
	insufficientPermissionsGeneric = "insufficient user permissions"
)

// JWTAuthMiddleware gets middleware that handles request authentication using
// a JWT bearer token.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := jwtAccessTokenValid(c); err != nil {
			logrus.Debug(err)
			c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
				ErrorMessage: authorizationFailedGeneric,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAllPermissionsMiddleware checks that the user making the request has
// all of the specified permissions.
func RequireAllPermissionsMiddleware(permissionKeys ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := JWTGetUser(c)
		if err != nil {
			logrus.Debug(err)
			c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
				ErrorMessage: authorizationFailedGeneric,
			})
			c.Abort()
			return
		}

		if u.Admin {
			c.Next()
			return
		}

		permissions, err := GetUserPermissions(c, u, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
				ErrorMessage: httperror.InternalServerError,
			})
			c.Abort()
			return
		}

		userPermissionKeys := map[string]struct{}{}
		for _, permission := range permissions {
			userPermissionKeys[permission.Key] = struct{}{}
		}

		for _, permissionKey := range permissionKeys {
			if _, ok := userPermissionKeys[permissionKey]; !ok {
				c.JSON(http.StatusForbidden, httperror.ErrorResponse{
					ErrorMessage: insufficientPermissionsGeneric,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequireAnyPermissionsMiddleware checks that the user making the request has
// at least one of the specified permissions.
func RequireAnyPermissionsMiddleware(permissionKeys ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, err := JWTGetUser(c)
		if err != nil {
			logrus.Debug(err)
			c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
				ErrorMessage: authorizationFailedGeneric,
			})
			c.Abort()
			return
		}

		if u.Admin {
			c.Next()
			return
		}

		permissions, err := GetUserPermissions(c, u, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
				ErrorMessage: httperror.InternalServerError,
			})
			c.Abort()
			return
		}

		userPermissionKeys := map[string]struct{}{}
		for _, permission := range permissions {
			userPermissionKeys[permission.Key] = struct{}{}
		}

		for _, permissionKey := range permissionKeys {
			if _, ok := userPermissionKeys[permissionKey]; ok {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, httperror.ErrorResponse{
			ErrorMessage: insufficientPermissionsGeneric,
		})
		c.Abort()
		return
	}
}

// JWTGetUser extracts a user record from the request access token.
func JWTGetUser(c *gin.Context) (*User, error) {

	metadata, err := jwtGetAccessMetadata(c)
	if err != nil {
		return nil, err
	}

	return GetUserByID(c, data.DB(), metadata.userID)

}

// JWTGetUserLogin extracts a user login record from the request access token.
func JWTGetUserLogin(c *gin.Context) (*Login, error) {

	metadata, err := jwtGetAccessMetadata(c)
	if err != nil {
		return nil, err
	}

	return GetLoginByUUID(c, data.DB(), metadata.authUUID)

}

// JWTValidateRefreshToken checks whether the supplied refresh token is valid,
// returns the associated user login record if the token is valid.
func JWTValidateRefreshToken(c *gin.Context,
	refreshToken string) (*Login, error) {

	metadata, err := jwtGetRefreshMetadata(c, refreshToken)
	if err != nil {
		return nil, err
	}

	login, err := GetLoginByUUID(c, data.DB(), metadata.authUUID)
	if err != nil {
		return nil, err
	}

	if metadata.expiresAt.Before(time.Now()) ||
		login.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	return login, nil

}

// jwtAccessTokenValid checks whether the request access token is valid.
func jwtAccessTokenValid(c *gin.Context) error {

	metadata, err := jwtGetAccessMetadata(c)
	if err != nil {
		return err
	}

	u, err := GetUserByID(c, data.DB(), metadata.userID)
	if err != nil {
		return err
	}

	if metadata.expiresAt.Before(time.Now()) ||
		(u.LoggedOutAt != nil && metadata.createdAt.Before(*u.LoggedOutAt)) {
		return errors.New("access token expired")
	}

	return nil

}

// jwtGetAccessMetadata extracts metdata from the request access token.
func jwtGetAccessMetadata(c *gin.Context) (*jwtAccessMetadata, error) {

	// parse JWT
	token, err := jwt.Parse(getAccessToken(c),
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v",
					token.Header["alg"])
			}
			return []byte(accessKey), nil
		})
	if err != nil {
		return nil, err
	}

	// define generic error to return return if parsing details fails
	genericErr := errors.New("failed to read JWT metadata")

	// extract claims from JWT
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, genericErr
	}

	authUUID, ok := claims["auth_uuid"].(string)
	if !ok {
		return nil, genericErr
	}

	userID, err := jwtParseIntFromClaims(claims, "user_id")
	if err != nil {
		return nil, genericErr
	}

	createdAtUnix, err := jwtParseIntFromClaims(claims, "created_at")
	if err != nil {
		return nil, genericErr
	}

	expiresAtUnix, err := jwtParseIntFromClaims(claims, "expires_at")
	if err != nil {
		return nil, genericErr
	}

	return &jwtAccessMetadata{
		authUUID:  authUUID,
		userID:    uint(userID),
		createdAt: time.Unix(int64(createdAtUnix), 0),
		expiresAt: time.Unix(int64(expiresAtUnix), 0),
	}, nil
}

// getAccessToken retrieves the bearer auth token from the supplied request.
func getAccessToken(c *gin.Context) string {

	tokenParts := strings.Split(c.Request.Header.Get("Authorization"), " ")

	if len(tokenParts) == 2 {
		return tokenParts[1]
	}

	return ""

}

// jwtGetRefreshMetadata extracts metdata from the supplied refresh token.
func jwtGetRefreshMetadata(c *gin.Context,
	refreshToken string) (*jwtRefreshMetadata, error) {

	// parse JWT
	token, err := jwt.Parse(refreshToken,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v",
					token.Header["alg"])
			}
			return []byte(refreshKey), nil
		})
	if err != nil {
		return nil, err
	}

	// define generic error to return return if parsing details fails
	genericErr := errors.New("failed to read JWT metadata")

	// extract claims from JWT
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, genericErr
	}

	authUUID, ok := claims["auth_uuid"].(string)
	if !ok {
		return nil, genericErr
	}

	userID, err := jwtParseIntFromClaims(claims, "user_id")
	if err != nil {
		return nil, genericErr
	}

	createdAtUnix, err := jwtParseIntFromClaims(claims, "created_at")
	if err != nil {
		return nil, genericErr
	}

	expiresAtUnix, err := jwtParseIntFromClaims(claims, "expires_at")
	if err != nil {
		return nil, genericErr
	}

	return &jwtRefreshMetadata{
		authUUID:  authUUID,
		userID:    uint(userID),
		createdAt: time.Unix(int64(createdAtUnix), 0),
		expiresAt: time.Unix(int64(expiresAtUnix), 0),
	}, nil
}

// jwtParseIntFromClaims extracts an integer from the supplied JWT map claims.
func jwtParseIntFromClaims(claims jwt.MapClaims, key string) (int, error) {

	var value int
	var err error

	switch claims[key].(type) {
	case string:
		value, err = strconv.Atoi(claims[key].(string))
	case float64:
		value = int(claims[key].(float64))
	default:
		return 0, fmt.Errorf("valid type for claim '%s'", key)
	}

	if err != nil {
		return 0, fmt.Errorf("invalid format for claim '%s'", key)
	}

	return value, nil

}
