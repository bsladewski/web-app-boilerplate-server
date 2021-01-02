package delivery

// signupRequest is used to read a request to the signup endpoint.
type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// signupEmailData is used to format the email sent by the signup endpoint.
type signupEmailData struct {
	ClientHost        string
	VerificationToken string
}

// signupVerifyRequest is used to read a request to the verify endpoint.
type signupVerifyRequest struct {
	Token string `json:"token"`
}

// loginRequest is used to read a request to the login endpoint.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse is used to format responses from the login endpoint.
type loginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Permissions  []string `json:"permissions"`
}

// refreshRequest is used to read a request to the refresh endpoint.
type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// refreshResponse is used to format responses from the refresh endpoint.
type refreshResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Permissions  []string `json:"permissions"`
}

// recoverRequest is used to read a request to the recover endpoint.
type recoverRequest struct {
	Email string `json:"email"`
}

// recoverEmailData is used to format the email sent by the recover endpoint.
type recoverEmailData struct {
	ClientHost        string
	VerificationToken string
}

// recoverResetRequest is used to read a request to reset a user account
// password as part of the account recovery process.
type recoverResetRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// resetRequest is used to read a request to reset the logged in user's account
// password
type resetRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
