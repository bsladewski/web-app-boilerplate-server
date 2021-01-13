package user

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"time"

	"web-app/data"
	"web-app/env"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"gorm.io/gorm"
)

// init configures the user package. This function reads an access and refresh
// key from the environment for JWT signing, if these keys are not found the
// application will log a fatal error.
func init() {

	// get access key for signing access tokens
	accessKey = env.MustGetString(accessKeyVariable)

	// get refresh key for signing refresh tokens
	refreshKey = env.MustGetString(refreshKeyVariable)

	// configure access token expiration time
	accessExpirationHours = time.Duration(
		env.GetIntSafe(accessExpirationHoursVariable, 1)) * time.Hour

	// configure refresh token expiration time
	refreshExpirationHours = time.Duration(
		env.GetIntSafe(refreshExpirationHoursVariable, 72)) * time.Hour

}

const (
	// accessKeyVariable defines an environment variable for the key used to
	// sign JWT access tokens.
	accessKeyVariable = "WEB_APP_ACCESS_KEY"
	// refreshKeyVariables defines an environment variable for the key used to
	// sign JWT refresh tokens.
	refreshKeyVariable = "WEB_APP_REFRESH_KEY"
	// accessExpirationHoursVariable defines an environment variable for the
	// number of hours before we should consider an access token expired.
	accessExpirationHoursVariable = "WEB_APP_ACCESS_EXPIRATION_HOURS"
	// refreshExpirationHoursVariable defines an environment variable for the
	// number of hours before we should consider a refresh token expired.
	refreshExpirationHoursVariable = "WEB_APP_REFRESH_EXPIRATION_HOURS"
)

// accessKey is used to sign JWT access tokens.
var accessKey string

// refreshKey is used to sign JWT refresh tokens.
var refreshKey string

// authExpirationHours determines the number of hours before we consider an
// access token to be expired.
var accessExpirationHours time.Duration

// refreshExpirationHours determines the number of hours before we consider a
// refresh token to be expired.
var refreshExpirationHours time.Duration

// CreateAuth generates JWT access and refresh tokens for the supplied user.
func CreateAuth(ctx context.Context, u *User) (accessToken,
	refreshToken string, err error) {

	// generate UUID to track issued credentials in peristent storage
	authUUID := uuid.NewV4().String()

	// create the access token
	accessJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"auth_uuid":  authUUID,
		"user_id":    u.ID,
		"created_at": time.Now().Unix(),
		"expires_at": time.Now().Add(accessExpirationHours).Unix(),
	})

	accessToken, err = accessJWT.SignedString([]byte(accessKey))
	if err != nil {
		return "", "", err
	}

	// create the refresh token
	refreshExpiration := time.Now().Add(refreshExpirationHours)

	refreshJWT := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"auth_uuid":  authUUID,
		"user_id":    u.ID,
		"created_at": time.Now().Unix(),
		"expires_at": refreshExpiration.Unix(),
	})

	refreshToken, err = refreshJWT.SignedString([]byte(refreshKey))
	if err != nil {
		return "", "", err
	}

	// add the user auth record
	if err := SaveLogin(ctx, data.DB(), &Login{
		UserID:    u.ID,
		UUID:      authUUID,
		ExpiresAt: refreshExpiration,
	}); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil

}

// GenerateSecretToken creates a base64 encoded token that includes both the
// supplied user id as well as the supplied payload encrypted with the user
// secret key.
func GenerateSecretToken(ctx context.Context, u *User,
	payload string) (string, error) {

	// create cipher with user secret key
	cipherBlock, err := aes.NewCipher([]byte(u.SecretKey))
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// encrypt and base64 encode payload
	payload = base64.URLEncoding.EncodeToString(aead.Seal(nonce, nonce,
		[]byte(payload), nil))

	// marshal token contents to json
	contents, err := json.Marshal(struct {
		UserID  uint
		Payload string
	}{
		UserID:  u.ID,
		Payload: payload,
	})
	if err != nil {
		return "", err
	}

	// base64 encode json token contents
	return base64.StdEncoding.EncodeToString(contents), nil

}

// ParseSecretToken parses the supplied secret token and returns the user id
// associated with the token as well as the decrypted payload string.
func ParseSecretToken(ctx context.Context,
	token string) (u *User, payload string, err error) {

	// base64 decode token contents
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, "", err
	}

	// unmarshal json token contents
	var tokenData = struct {
		UserID  uint
		Payload string
	}{}

	if err = json.Unmarshal(tokenBytes, &tokenData); err != nil {
		return nil, "", err
	}

	// get user record
	u, err = GetUserByID(ctx, data.DB(), tokenData.UserID)
	if err != nil {
		return nil, "", err
	}

	// base64 decode encrypted payload
	encryptData, err := base64.URLEncoding.DecodeString(tokenData.Payload)
	if err != nil {
		return nil, "", err
	}

	// create cipher with user secret key
	cipherBlock, err := aes.NewCipher([]byte(u.SecretKey))
	if err != nil {
		return nil, "", err
	}

	aead, err := cipher.NewGCM(cipherBlock)
	if err != nil {
		return nil, "", err
	}

	nonceSize := aead.NonceSize()
	if len(encryptData) < nonceSize {
		return nil, "", err
	}

	// decrypt the payload
	nonce, cipherText := encryptData[:nonceSize], encryptData[nonceSize:]
	payloadBytes, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, "", err
	}

	// return string representation of payload
	return u, string(payloadBytes), nil

}

// GetUserPermissions returns a list of permissions associated with the supplied
// user and the user's assigned roles.
func GetUserPermissions(ctx context.Context, u *User,
	public *bool) ([]*Permission, error) {

	// if the user is marked as an admin return all permissions
	if u.Admin {
		return ListPermission(ctx, data.DB(), public)
	}

	// retrieve permissions directly associated with the user
	results, err := ListPermissionByUser(ctx, data.DB(), u.ID, public)
	if err != nil {
		return nil, err
	}

	// retrieve roles associated with the user
	roles, err := ListRoleByUser(ctx, data.DB(), u.ID)
	if err != nil {
		return nil, err
	}

	// keep track of permissions we have already added
	added := map[string]struct{}{}
	for _, permission := range results {
		added[permission.Key] = struct{}{}
	}

	// retrieve permissions associated with the user roles
	for _, role := range roles {
		permissions, err := ListPermissionByRole(ctx, data.DB(), role.ID, public)
		if err != nil {
			return nil, err
		}

		for i := 0; i < len(permissions); i++ {
			if _, ok := added[permissions[i].Key]; !ok {
				results = append(results, permissions[i])
				added[permissions[i].Key] = struct{}{}
			}
		}
	}

	return results, nil

}

// CreateRoles will create all specified roles if they do not already exist.
func CreateRoles(ctx context.Context, roleKeys ...string) error {

	for _, roleKey := range roleKeys {
		if err := CreateRole(ctx, roleKey); err != nil {
			return err
		}
	}

	return nil

}

// CreateRole will create the specified role if it does not already exist. If
// the role already exists nothing will happen and no error will be returned.
func CreateRole(ctx context.Context, roleKey string) error {

	// check if the role already exists
	_, err := GetRoleByKey(ctx, data.DB(), roleKey)
	if err != gorm.ErrRecordNotFound {
		return err
	}

	// if the role does not exist create it
	return SaveRole(ctx, data.DB(), &Role{
		ReadOnly: true,
		Key:      roleKey,
	})

}

// CreatePublicPermissions will create all specified permissions if they do
// not already exist. The roles will be marked as public pemrissions. If a list
// of roles is supplied all permissions will also be associated with the list of
// roles.
func CreatePublicPermissions(ctx context.Context, permissionKeys []string,
	roles []string) error {

	for _, permissionKey := range permissionKeys {
		if err := CreatePermission(ctx, permissionKey, true, roles...); err != nil {
			return err
		}
	}

	return nil

}

// CreatePrivatePermissions will create all specified permissions if they do
// not already exist. The roles will be marked as private pemrissions. If a list
// of roles is supplied all permissions will also be associated with the list of
// roles.
func CreatePrivatePermissions(ctx context.Context, permissionKeys []string,
	roles []string) error {

	for _, permissionKey := range permissionKeys {
		if err := CreatePermission(ctx, permissionKey, false, roles...); err != nil {
			return err
		}
	}

	return nil

}

// CreatePermission will create the specified permission if it does not already
// exist. Additionally, the permission will be associated with the supplied list
// of roles. If the permission already exists nothing will happen and no error
// will be returned.
func CreatePermission(ctx context.Context, permissionKey string, public bool,
	roles ...string) error {

	// create a new transaction
	tx := data.DB().Begin()

	// wrap the work in a function to capture any errors and simplify committing
	// or rolling back the transaction
	if err := func() error {

		// check if the permission already exists
		permission, err := GetPermissionByKey(ctx, tx, permissionKey)
		if err == gorm.ErrRecordNotFound {
			permission = &Permission{
				Public: public,
				Key:    permissionKey,
			}

			// if the permission does not already exist create it
			if err := SavePermission(ctx, tx, permission); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		// create associated between the permission and roles that should use
		// the permission
		for _, roleKey := range roles {
			role, err := GetRoleByKey(ctx, tx, roleKey)
			if err != nil {
				return err
			}

			if err := saveRolePermission(ctx, tx, &rolePermission{
				RoleID:       role.ID,
				PermissionID: permission.ID,
			}); err != nil {
				return err
			}
		}

		return nil

	}(); err != nil {
		// if an error was encountered roll back the transaction
		if err := tx.Rollback(); err != nil {
			logrus.Error(err)
		}
		return err
	}

	// if no error was encountered commit the transaction
	return tx.Commit().Error

}
