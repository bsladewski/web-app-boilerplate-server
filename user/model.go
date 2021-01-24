package user

import (
	"time"

	"web-app/data"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// init migrates the database model.
func init() {
	data.DB().AutoMigrate(
		User{},
		Login{},
		Role{},
		Permission{},
		userRole{},
		rolePermission{},
		userPermission{},
	)

	// check if we should use mock data
	if !data.UseMockData() {
		return
	}

	// load mock data
	for _, u := range mockUsers {
		if err := data.DB().Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&u).Error; err != nil {
			logrus.Fatal(err)
		}
	}
}

/* Data Types */

// User provides access to the application.
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Email    string `gorm:"index,unique" json:"email"`
	Password string `json:"password"`

	Admin     bool   `json:"admin"`      // admins have the broadest set of user permissions
	SecretKey string `json:"secret_key"` // used to sign tokens when generating links for this user
	Verified  bool   `json:"verified"`   // whether the user has completed email verification

	LoggedOutAt *time.Time `json:"logged_out_at"` // records the last time the user explicitly logged out
}

// Login stores identifiers for validating user auth tokens.
type Login struct {
	ID uint `gorm:"primarykey" json:"id"`

	UserID uint   `gorm:"index" json:"user_id"`
	UUID   string `gorm:"index" json:"uuid"` // uniquely identifies a refresh token

	ExpiresAt time.Time `json:"expires_at"` // records when a refresh token will expire
}

// Role represents a predifined set of permissions that may be applied to a
// user.
type Role struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	ReadOnly    bool   `gorm:"index" json:"read_only"`  // read only permissions cannot be edited or deleted
	Key         string `gorm:"index,unique" json:"key"` // text that uniquely identifies this role
	Name        string `json:"name"`                    // a display name for this role
	Description string `json:"description"`             // a brief description of this role
}

// userRole relates a user account to a role.
type userRole struct {
	ID uint `gorm:"primarykey" json:"id"`

	UserID uint `json:"user_id"`
	User   User `gorm:"constraint:OnDelete:CASCADE"`
	RoleID uint `json:"role_id"`
	Role   Role `gorm:"constraint:OnDelete:CASCADE"`
}

// Permission allows a user to access some feature of the application.
type Permission struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Public      bool   `gorm:"index" json:"public"`     // public permissions are also used on the front-end
	Key         string `gorm:"index,unique" json:"key"` // text that uniquely identifies this permission
	Name        string `json:"name"`                    // a display name for this permission
	Description string `json:"description"`             // a brief description of this permission
}

// rolePermission relates a role to a permission.
type rolePermission struct {
	ID uint `gorm:"primarykey" json:"id"`

	RoleID       uint       `json:"role_id"`
	Role         Role       `gorm:"constraint:OnDelete:CASCADE"`
	PermissionID uint       `json:"permission_id"`
	Permission   Permission `gorm:"constraint:OnDelete:CASCADE"`
}

// userPermission relates a user to a permission.
type userPermission struct {
	ID uint `gorm:"primarykey" json:"id"`

	UserID       uint       `json:"user_id"`
	User         User       `gorm:"constraint:OnDelete:CASCADE"`
	PermissionID uint       `json:"permission_id"`
	Permission   Permission `gorm:"constraint:OnDelete:CASCADE"`
}

/* Mock Data */

var mockUsers = []User{
	{
		ID:        1,
		Email:     "admin@example.com",
		Password:  "$2a$10$JrF8wIP/MaN.i5xx5VV.ZuqP3DxJs1Q4fAY.WbGPOhYWQzpon3kpm", // pass_good
		Admin:     true,
		SecretKey: "8bf83c80-f235-461e-9bd7-00c83a5cfff8",
		Verified:  true,
	},
	{
		ID:        2,
		Email:     "test@example.com",
		Password:  "$2a$10$JrF8wIP/MaN.i5xx5VV.ZuqP3DxJs1Q4fAY.WbGPOhYWQzpon3kpm", // pass_good
		SecretKey: "43ee0e83-dc81-4263-8bb0-6ccddff8586d",
		Verified:  true,
	},
}
