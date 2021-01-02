package user

import (
	"time"

	"github.com/bsladewski/web-app-boilerplate-server/data"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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
		if err := data.DB().Create(&u).Error; err != nil {
			logrus.Fatal(err)
		}
	}

	for _, r := range mockRoles {
		if err := data.DB().Create(&r).Error; err != nil {
			logrus.Fatal(err)
		}
	}

	for _, ur := range mockUserRoles {
		if err := data.DB().Create(&ur).Error; err != nil {
			logrus.Fatal(err)
		}
	}

	for _, p := range mockPermissions {
		if err := data.DB().Create(&p).Error; err != nil {
			logrus.Fatal(err)
		}
	}

	for _, up := range mockUserPermissions {
		if err := data.DB().Create(&up).Error; err != nil {
			logrus.Fatal(err)
		}
	}

	for _, rp := range mockRolePermissions {
		if err := data.DB().Create(&rp).Error; err != nil {
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

	LoggedOutAt time.Time `json:"logged_out_at"` // records the last time the user explicitly logged out
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

var mockRoles = []Role{
	{
		ID:          1,
		Key:         "example_role_1",
		Name:        "Example Role 1",
		Description: "Adds example public permissions.",
	},
	{
		ID:          2,
		Key:         "example_role_2",
		Name:        "Example Role 2",
		Description: "Adds example private permissions.",
	},
}

var mockUserRoles = []userRole{
	{
		ID:     1,
		UserID: 2, // test@example.com
		RoleID: 1, // example_role_1
	},
	{
		ID:     2,
		UserID: 2, // test@example.com
		RoleID: 2, // example_role_2
	},
}

var mockPermissions = []Permission{
	{
		ID:          1,
		Key:         "example_direct_1",
		Name:        "Example Direct 1",
		Description: "This permission will be assigned directly to a user.",
	},
	{
		ID:          2,
		Key:         "example_direct_2",
		Name:        "Example Direct 2",
		Description: "This permission will be assigned directly to a user.",
	},
	{
		ID:          3,
		Public:      true,
		Key:         "example_public_1",
		Name:        "Example Public 1",
		Description: "This permission will be used by both the front-end and the back-end.",
	},
	{
		ID:          4,
		Public:      true,
		Key:         "example_public_2",
		Name:        "Example Public 2",
		Description: "This permission will be used by both the front-end and the back-end.",
	},
	{
		ID:          5,
		Key:         "example_private_1",
		Name:        "Example Private 1",
		Description: "This permission will only be used by the back-end.",
	},
	{
		ID:          6,
		Key:         "example_private_2",
		Name:        "Example Private 2",
		Description: "This permission will only be used by the back-end.",
	},
}

var mockUserPermissions = []userPermission{
	{
		ID:           1,
		UserID:       2, // test@example.com
		PermissionID: 1, // example_direct_1
	},
	{
		ID:           2,
		UserID:       2, // test@example.com
		PermissionID: 2, // example_direct_2
	},
}

var mockRolePermissions = []rolePermission{
	{
		ID:           1,
		RoleID:       1, // example_role_1
		PermissionID: 3, // example_public_1
	},
	{
		ID:           2,
		RoleID:       1, // example_role_1
		PermissionID: 4, // example_public_2
	},
	{
		ID:           3,
		RoleID:       2, // example_role_2
		PermissionID: 5, // example_private_1
	},
	{
		ID:           4,
		RoleID:       2, // example_role_2
		PermissionID: 6, // example_private_2
	},
}
