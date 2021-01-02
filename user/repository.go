package user

import (
	"context"
	"time"

	"gorm.io/gorm"
)

////////////////////////////////////////////////////////////////////////////////
// User                                                                       //
////////////////////////////////////////////////////////////////////////////////

// GetUserByID retrieves a user record by id.
func GetUserByID(ctx context.Context, db *gorm.DB, id uint) (*User, error) {

	var item User

	if err := db.Model(&User{}).
		Where("id = ?", id).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// GetUserByEmail retrieves a user record by email address.
func GetUserByEmail(ctx context.Context, db *gorm.DB,
	email string) (*User, error) {

	var item User

	if err := db.Model(&User{}).
		Where("LOWER(email) = LOWER(?)", email).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// SaveUser inserts or updates the supplied user record.
func SaveUser(ctx context.Context, db *gorm.DB, item *User) error {
	return db.Save(item).Error
}

// DeleteUser deletes the supplied user record.
func DeleteUser(ctx context.Context, db *gorm.DB, item *User) error {
	return db.Delete(item).Error
}

////////////////////////////////////////////////////////////////////////////////
// Login                                                                      //
////////////////////////////////////////////////////////////////////////////////

// GetLoginByID retrieves a user login record by id.
func GetLoginByID(ctx context.Context, db *gorm.DB,
	id uint) (*Login, error) {

	var item Login

	if err := db.Model(&Login{}).
		Where("id = ?", id).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// GetLoginByUUID retrieves a user login record by UUID.
func GetLoginByUUID(ctx context.Context, db *gorm.DB,
	uuid string) (*Login, error) {

	var item Login

	if err := db.Model(&Login{}).
		Where("uuid = ?", uuid).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// ListLoginByUserID retrieves all user login records associated with the
// supplied user id.
func ListLoginByUserID(ctx context.Context, db *gorm.DB,
	userID uint) ([]*Login, error) {

	var items []*Login

	if err := db.Model(&Login{}).
		Where("user_id = ?", userID).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil

}

// SaveLogin inserts or updates the supplied user login record.
func SaveLogin(ctx context.Context, db *gorm.DB, item *Login) error {
	return db.Save(item).Error
}

// DeleteLogin deletes the supplied user login record.
func DeleteLogin(ctx context.Context, db *gorm.DB, item *Login) error {
	return db.Delete(item).Error
}

// DeleteExpiredLogin deletes all expires user login records associated with
// the specified user id.
func DeleteExpiredLogin(ctx context.Context, db *gorm.DB, userID uint) error {
	return db.
		Where("user_id = ?", userID).
		Where("expires_at < ?", time.Now()).
		Delete(&Login{}).Error
}

////////////////////////////////////////////////////////////////////////////////
// Role                                                                       //
////////////////////////////////////////////////////////////////////////////////

// GetRoleByID retrieves a role by id.
func GetRoleByID(ctx context.Context, db *gorm.DB, id uint) (*Role, error) {

	var item Role

	if err := db.Model(&Role{}).
		Where("id = ?", id).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// GetRoleByKey retrieves a role by its key.
func GetRoleByKey(ctx context.Context, db *gorm.DB, key string) (*Role, error) {

	var item Role

	if err := db.Model(&Role{}).
		Where("key = ?", key).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// ListRole retrieves all defined roles.
func ListRole(ctx context.Context, db *gorm.DB) ([]*Role, error) {

	var items []*Role

	if err := db.Model(&Role{}).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil

}

// ListRoleByUser retrieves all roles associated with the specified user.
func ListRoleByUser(ctx context.Context, db *gorm.DB,
	userID uint) ([]*Role, error) {

	var items []*userRole

	if err := db.Model(&userRole{}).
		Where("user_id = ?", userID).
		Preload("Role").Find(&items).Error; err != nil {
		return nil, err
	}

	var roles []*Role

	for _, item := range items {
		roles = append(roles, &item.Role)
	}

	return roles, nil

}

// SaveRole inserts or updates the supplied role record.
func SaveRole(ctx context.Context, db *gorm.DB, item *Role) error {
	return db.Save(item).Error
}

// DeleteRole deletes the supplied role record.
func DeleteRole(ctx context.Context, db *gorm.DB, item *Role) error {
	return db.Delete(item).Error
}

////////////////////////////////////////////////////////////////////////////////
// Permission                                                                 //
////////////////////////////////////////////////////////////////////////////////

// GetPermissionByID retrieves a permission by id.
func GetPermissionByID(ctx context.Context, db *gorm.DB,
	id uint) (*Permission, error) {

	var item Permission

	if err := db.Model(&Permission{}).
		Where("id = ?", id).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// GetPermissionByKey retrieves a permission by its key.
func GetPermissionByKey(ctx context.Context, db *gorm.DB,
	key string) (*Permission, error) {

	var item Permission

	if err := db.Model(&Permission{}).
		Where("key = ?", key).
		First(&item).Error; err != nil {
		return nil, err
	}

	return &item, nil

}

// ListPermission retrieves all defined permissions. May optionally be filtered
// by whether the permission is public.
func ListPermission(ctx context.Context, db *gorm.DB,
	public *bool) ([]*Permission, error) {

	var items []*Permission

	q := db.Model(&Permission{})

	if public != nil {
		q = q.Where("public = ?", *public)
	}

	if err := q.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil

}

// ListPermissionByRole retrieves all permissions associated with the specified
// role. May optionally be filtered by whether the permission is public.
func ListPermissionByRole(ctx context.Context, db *gorm.DB,
	roleID uint, public *bool) ([]*Permission, error) {

	var items []*rolePermission

	q := db.Model(&rolePermission{}).
		Where("role_id = ?", roleID)

	if err := q.Preload("Permission").Find(&items).Error; err != nil {
		return nil, err
	}

	var permissions []*Permission

	for _, item := range items {
		if public != nil && item.Permission.Public != *public {
			continue
		}
		permissions = append(permissions, &item.Permission)
	}

	return permissions, nil

}

// ListPermissionByUser retrieves all permissions associated with the specified
// user. May optionally be filtered by whether the permission is public.
func ListPermissionByUser(ctx context.Context, db *gorm.DB,
	userID uint, public *bool) ([]*Permission, error) {

	var items []*userPermission

	q := db.Model(&userPermission{}).
		Where("user_id = ?", userID)

	if err := q.Preload("Permission").Find(&items).Error; err != nil {
		return nil, err
	}

	var permissions []*Permission

	for _, item := range items {
		if public != nil && item.Permission.Public != *public {
			continue
		}
		permissions = append(permissions, &item.Permission)
	}

	return permissions, nil

}

// SavePermission inserts or updates the supplied permission record.
func SavePermission(ctx context.Context, db *gorm.DB, item *Permission) error {
	return db.Save(item).Error
}

// DeletePermission deletes the supplied permission record.
func DeletePermission(ctx context.Context, db *gorm.DB,
	item *Permission) error {
	return db.Delete(item).Error
}
