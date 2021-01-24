package delivery

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"web-app/data"
	"web-app/email"
	"web-app/httperror"
	"web-app/server"
	"web-app/user"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// init registers the user API with the application router.
func init() {

	// bind public endpoints
	server.Router().POST(signupEndpoint, signup)
	server.Router().POST(signupVerifyEndpoint, signupVerify)
	server.Router().POST(loginEndpoint, login)
	server.Router().POST(refreshEndpoint, refresh)
	server.Router().POST(recoverEndpoint, recover)
	server.Router().POST(recoverResetEndpoint, recoverReset)

	// bind private endpoints
	server.Router().POST(logoutEndpoint, user.JWTAuthMiddleware(), logout)
	server.Router().POST(resetEndpoint, user.JWTAuthMiddleware(), reset)
}

const (
	// signupEndpoint the API endpoint used to create new user accounts.
	signupEndpoint = "/signup"
	// signupVerifyEndpoint the API endpoint used to verify a new user's email
	// address.
	signupVerifyEndpoint = "/signup/verify"
	// loginEndpoint the API endpoint that handles user login.
	loginEndpoint = "/login"
	// refreshEndpoint the API endpoint that handles refreshing access tokens.
	refreshEndpoint = "/refresh"
	// logoutEndpoint the API endpoint that handles user logout.
	logoutEndpoint = "/logout"
	// recoverEndpoint the API endpoint used to send account recovery emails.
	recoverEndpoint = "/recover"
	// recoverResetEndpoint the API endpoint for resetting an account password
	// as part of the account recovery process.
	recoverResetEndpoint = "/recover/reset"
	// resetEndpoint the API endpoint used to reset the logged in user's
	// password.
	resetEndpoint = "/reset"
	// invalidToken is an error returned if if a user validation token is
	// supplied that cannot be parsed or contains invalid data.
	invalidToken = "invalid token"
	// resetFailedGeneric is a generic error message returned when resetting
	// the user account password fails.
	resetFailedGeneric = "failed to reset password"
	// invalidUserCredentials is an error message returned when the user's email
	// or password is incorrect.
	invalidUserCredentials = "invalid email or password"
	// logoutFailedGeneric is a generic error returned when user logout fails.
	logoutFailedGeneric = "failed to log out user"
	// invalidRefreshToken is an error message returned if the user supplies an
	// invalid refresh token or a refresh token that is inconsistent with
	// persistent data.
	invalidRefreshToken = "invalid refresh token"
)

// signup creates a new user account.
func signup(c *gin.Context) {

	var req signupRequest

	// read request parameters
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "invalid request body",
		})
		return
	}

	// validate request parameters
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "email is required",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "password is required",
		})
		return
	}

	// check if a verified user account with the same email address already
	// exists
	u, err := user.GetUserByEmail(c, data.DB(), req.Email)
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	} else if u != nil && u.Verified {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "email address is already registered",
		})
		return
	}

	// create a new transaction for creating the user account
	tx := data.DB().Begin()

	// if no unverified user account exists, create a new user account
	if u == nil {

		// generate user secret key
		secretKey := md5.Sum(uuid.NewV4().Bytes())

		u = &user.User{
			Email:     req.Email,
			SecretKey: fmt.Sprintf("%x", secretKey),
		}

		// create the user account record
		if err := user.SaveUser(c, tx, u); err != nil {
			logrus.Error(err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
				ErrorMessage: httperror.InternalServerError,
			})
			return
		}

	}

	// set user password
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(fmt.Sprintf("%d:%s", u.ID, req.Password)), bcrypt.DefaultCost)
	if err != nil {
		logrus.Error(err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	u.Password = string(hash)

	if err := user.SaveUser(c, tx, u); err != nil {
		logrus.Error(err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	// generate the verification token
	token, err := user.GenerateSecretToken(c, u, u.Email)
	if err != nil {
		logrus.Error(err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	// send the verification email
	if err := email.SendEmailTemplate(
		email.DefaultFromAddress(),
		email.DefaultReplyToAddress(),
		[]string{u.Email},
		nil,
		nil,
		email.TemplateTitleSignup,
		signupEmailData{
			ClientHost:        server.ClientBaseURL(),
			VerificationToken: token,
		},
	); err != nil {
		logrus.Error(err)
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: "failed to send verification email, please try again later",
		})
		return
	}

	// commit the transaction
	if err := tx.Commit().Error; err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: "failed to create user account, please try again later",
		})
		return
	}

}

// signupVerify checks the supplied verification token to determine if the user
// has access to the account email address.
func signupVerify(c *gin.Context) {

	var req signupVerifyRequest

	// read request parameters
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "invalid request body",
		})
		return
	}

	// decode the verification token
	u, payload, err := user.ParseSecretToken(c, req.Token)
	if err != nil {
		logrus.Warn(err)
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: invalidToken,
		})
		return
	}

	// validate the token payload
	if payload != u.Email {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: invalidToken,
		})
		return
	}

	// mark the user record as verified
	u.Verified = true

	// save user record
	if err := user.SaveUser(c, data.DB(), u); err != nil {
		logrus.WithError(err)
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: invalidToken,
		})
		return
	}

	// response with 200 - OK if verification was successful
	c.Status(http.StatusOK)

}

// login checks user credentials and generates access and refresh tokens for
// authenticating user requests.
func login(c *gin.Context) {

	var req loginRequest

	// read user credentials from request body
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "invalid request body",
		})
		return
	}

	// validate request parameters
	if req.Email == "" {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "email is required",
		})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "password is required",
		})
		return
	}

	// retrieve user account by email address
	u, err := user.GetUserByEmail(c, data.DB(), req.Email)
	if err == gorm.ErrRecordNotFound {
		logrus.Warn(err)
		c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
			ErrorMessage: invalidUserCredentials,
		})
		return
	} else if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	if !u.Verified {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "account email has not been verified",
		})
		return
	}

	// compare supplied password with user password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(fmt.Sprintf("%d:%s", u.ID, req.Password)),
	); err != nil {
		logrus.Debug(err)
		c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
			ErrorMessage: invalidUserCredentials,
		})
		return
	}

	// generate access and refresh tokens
	accessToken, refreshToken, err := user.CreateAuth(c, u)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	var permissionKeys []string

	// get public user permissions
	permissions, err := user.GetUserPermissions(c, u, ptrToBool(true))
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	for _, permission := range permissions {
		permissionKeys = append(permissionKeys, permission.Key)
	}

	// delete expired user login records to keep persistent storage clean
	go func() {
		if err := user.DeleteExpiredLogin(c, data.DB(), u.ID); err != nil {
			logrus.Error(err)
		}
	}()

	// repond with auth tokens
	c.JSON(http.StatusOK, loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Permissions:  permissionKeys,
	})

}

// refresh checks the supplied refresh token and generates new access and
// refresh tokens if valid.
func refresh(c *gin.Context) {

	var req refreshRequest

	// read user credentials from request body
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "invalid request body",
		})
		return
	}

	// validate request parameters
	if req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "refresh token is required",
		})
		return
	}

	// validate the supplied refresh token
	login, err := user.JWTValidateRefreshToken(c, req.RefreshToken)
	if err != nil {
		logrus.Warn(err)
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: invalidRefreshToken,
		})
		return
	}

	// retrieve user record
	u, err := user.GetUserByID(c, data.DB(), login.UserID)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: invalidRefreshToken,
		})
		return
	}

	// generate access and refresh tokens
	accessToken, refreshToken, err := user.CreateAuth(c, u)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	// delete original refresh token
	if err := user.DeleteLogin(c, data.DB(), login); err != nil {
		logrus.Error(err)
	}

	var permissionKeys []string

	// get public user permissions
	permissions, err := user.GetUserPermissions(c, u, ptrToBool(true))
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	for _, permission := range permissions {
		permissionKeys = append(permissionKeys, permission.Key)
	}

	// repond with auth tokens
	c.JSON(http.StatusOK, refreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Permissions:  permissionKeys,
	})

}

// logout invalidates the logged in user's access and refresh tokens.
func logout(c *gin.Context) {

	// get user from JWT
	u, err := user.JWTGetUser(c)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
			ErrorMessage: logoutFailedGeneric,
		})
		return
	}

	// get user auth record from JWT
	login, err := user.JWTGetUserLogin(c)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
			ErrorMessage: logoutFailedGeneric,
		})
		return
	}

	// delete user auth record, this will invalidate the refresh token
	if err := user.DeleteLogin(c, data.DB(), login); err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: logoutFailedGeneric,
		})
		return
	}

	// set logged out at time, this will invalidate all access tokens issued
	// before this time
	now := time.Now()
	u.LoggedOutAt = &now

	// update the user record
	if err := user.SaveUser(c, data.DB(), u); err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: logoutFailedGeneric,
		})
		return
	}

	// respond with 200 - OK if logout was successful
	c.Status(http.StatusOK)

}

// recover sends an email to the user with a link to reset the user account
// password.
func recover(c *gin.Context) {

	var req recoverRequest

	// read request parameters
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "invalid request body",
		})
		return
	}

	// retrieve user account by email address
	u, err := user.GetUserByEmail(c, data.DB(), req.Email)
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "email address not found",
		})
		return
	} else if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	// generate the verification token
	token, err := user.GenerateSecretToken(c, u, u.Email)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	// send the verification email
	if err := email.SendEmailTemplate(
		email.DefaultFromAddress(),
		email.DefaultReplyToAddress(),
		[]string{u.Email},
		nil,
		nil,
		email.TemplateTitleRecover,
		recoverEmailData{
			ClientHost:        server.ClientBaseURL(),
			VerificationToken: token,
		},
	); err != nil {
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: "failed to send verification email, please try again later",
		})
	}

}

// recoverReset is used to change a user account password as part of the account
// recovery process.
func recoverReset(c *gin.Context) {

	var req recoverResetRequest

	// read request parameters
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "invalid request body",
		})
		return
	}

	// decode the verification token
	u, payload, err := user.ParseSecretToken(c, req.Token)
	if err != nil {
		logrus.Warn(err)
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: invalidToken,
		})
		return
	}

	// validate the token payload
	if payload != u.Email {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: invalidToken,
		})
		return
	}

	// set user password
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(fmt.Sprintf("%d:%s", u.ID, req.Password)), bcrypt.DefaultCost)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	u.Password = string(hash)

	if err := user.SaveUser(c, data.DB(), u); err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	// response with 200 - OK if password reset was successful
	c.Status(http.StatusOK)

}

// reset is used to change the logged in user's account password.
func reset(c *gin.Context) {

	// get user from JWT
	u, err := user.JWTGetUser(c)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusUnauthorized, httperror.ErrorResponse{
			ErrorMessage: resetFailedGeneric,
		})
		return
	}

	var req resetRequest

	// read request parameters
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "invalid request body",
		})
		return
	}

	// validate request parameters
	if req.CurrentPassword == "" {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "current password is required",
		})
		return
	}

	if req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "new password is required",
		})
		return
	}

	// verify current password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(u.Password),
		[]byte(fmt.Sprintf("%d:%s", u.ID, req.CurrentPassword)),
	); err != nil {
		logrus.Debug(err)
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "current password is incorrect",
		})
		return
	}

	// check that current password is not the same as the new password
	if req.CurrentPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, httperror.ErrorResponse{
			ErrorMessage: "new and current passwords are the same",
		})
		return
	}

	// set user password
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(fmt.Sprintf("%d:%s", u.ID, req.NewPassword)), bcrypt.DefaultCost)
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	u.Password = string(hash)

	if err := user.SaveUser(c, data.DB(), u); err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, httperror.ErrorResponse{
			ErrorMessage: httperror.InternalServerError,
		})
		return
	}

	// response with 200 - OK if password reset was successful
	c.Status(http.StatusOK)

}

// ptrToBool gets a pointer to the supplied boolean value.
func ptrToBool(val bool) *bool {
	return &val
}
