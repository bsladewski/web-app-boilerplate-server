// Package user provides functionality for managing user accounts, permissions,
// and authentication.
//
// Environment:
//     WEB_APP_ACCESS_KEY:
//         string - the key used to sign JWT access tokens
//     WEB_APP_REFRESH_KEY:
//         string - the key used to sign JWT refresh tokens
//     WEB_APP_ACCESS_EXPIRATION_HOURS:
//         int - the number of hours before an access token is expired
//         Default: 1
//     WEB_APP_REFRESH_EXPIRATION_HOURS:
//         int - the number of hours before a refresh token is expired
//         Default: 72
package user
