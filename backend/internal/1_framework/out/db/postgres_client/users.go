package postgres_client

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"backend/internal/1_framework/out/db/postgres_client/models"
	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
	"backend/internal/logger"
)

func (receiver *PostgresClient) ReplaceUser(
	ctx context.Context,
	name string,
	email string,
	userID string,
) (
	err error,
) {
	err = receiver.conn(ctx).Transaction(func(tx *gorm.DB) error {
		err := receiver.AddUser(tx, name, email)
		if err != nil {
			return err
		}
		err = receiver.DeleteUser(tx, userID)
		if err != nil {
			return err
		}

		return nil
	})

	return
}

func (receiver *PostgresClient) AddUser(
	tx *gorm.DB,
	name string,
	email string,
) (
	err error,
) {
	record := models.User{
		FullName: sql.NullString{String: name, Valid: name != ""},
		Email:    email,
	}

	err = tx.
		Omit("Auth0UserID", "CreatedAt", "UpdatedAt").
		Create(&record).
		Error

	return
}

func (receiver *PostgresClient) DeleteUser(
	tx *gorm.DB,
	userID string,
) (
	err error,
) {
	err = tx.Delete(&models.User{}, userID).Error

	return
}

func (receiver *PostgresClient) GetUserList(
	ctx context.Context,
) (
	userList groupObject.UserList,
	err error,
) {
	userList = groupObject.UserList{}
	err = nil
	users := []models.User{} // SQL結果保存用

	result := receiver.conn(ctx).
		Model(&models.User{}).
		Select("id", "email", "full_name").
		Find(&users)

	if result.Error != nil {
		err = result.Error
		return
	}

	if result.RowsAffected == 0 {
		return
	}

	userArgs := make([]groupObject.NewUserArgs, 0, len(users))
	for _, user := range users {
		userArgs = append(userArgs, groupObject.NewUserArgs{
			ID:    &user.ID,
			Name:  stringFromNullString(user.FullName),
			Email: &user.Email,
		})
	}

	userList, err = groupObject.ReconstructUserList(&groupObject.NewUserListArgs{
		Content: userArgs,
	})
	return
}

func (receiver *PostgresClient) GetUser(
	ctx context.Context,
	id typeObject.ID,
) (
	user groupObject.User,
	err error,
) {
	user = groupObject.User{}   // ドメインロジック用
	resultUser := models.User{} // SQL結果保存用

	result := receiver.conn(ctx).
		Model(&models.User{}).
		Select("id", "email", "full_name").
		Where(&models.User{ID: id.GetValue()}).
		Take(&resultUser)

	if result.Error != nil {
		err = result.Error
		return
	}

	args := &groupObject.NewUserArgs{
		ID:    &resultUser.ID,
		Name:  stringFromNullString(resultUser.FullName),
		Email: &resultUser.Email,
	}
	newUser, err := groupObject.ReconstructUser(args)
	if err != nil {
		return
	}

	user, err = *newUser, nil
	return
}

func (receiver *PostgresClient) CreateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	createdUser groupObject.User,
	err error,
) {
	if err = newUser.EnsureReadyToCreate(); err != nil {
		return
	}

	record := models.User{
		Email: newUser.Email().GetValue(),
		FullName: sql.NullString{
			String: newUser.Name().GetValue(),
			Valid:  true,
		},
	}
	if err = receiver.conn(ctx).
		Omit("Auth0UserID", "CreatedAt", "UpdatedAt").
		Create(&record).
		Error; err != nil {
		return
	}

	createdUserPointer, err := groupObject.ReconstructUser(&groupObject.NewUserArgs{
		ID:    &record.ID,
		Name:  stringFromNullString(record.FullName),
		Email: &record.Email,
	})
	if err != nil {
		return createdUser, err
	}

	createdUser = *createdUserPointer

	return
}

func (receiver *PostgresClient) UpdateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	err error,
) {
	if returnedErr := newUser.EnsureReadyToUpdate(); returnedErr != nil {
		err = returnedErr
		return
	}

	record := models.User{
		FullName: sql.NullString{
			String: newUser.Name().GetValue(),
			Valid:  true,
		},
		Email: newUser.Email().GetValue(),
	}

	result := receiver.conn(ctx).
		Model(&models.User{}).
		Where(&models.User{ID: newUser.Identity().GetValue()}).
		Select("full_name", "email").
		Updates(&record)
	if result.Error != nil {
		err = result.Error
		return
	}
	if result.RowsAffected == 0 {
		err = gorm.ErrRecordNotFound
		return
	}

	err = nil

	return
}

// GetUserListByCondition ...
func (receiver *PostgresClient) GetUserListByCondition(
	ctx context.Context,
	reqUser groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	resUserList = groupObject.UserList{}
	// logger.Logging(
	// 	ctx,
	// 	middlewareRequestContext.GetRequestContext(ctx).TraceID.GetValue(),
	// )

	users := []models.User{} // SQL結果保存用

	conn := receiver.conn(ctx).
		Model(&models.User{}).
		Select("id", "email", "full_name")

	if !reqUser.Email().GetIsNil() && reqUser.Email().GetValue() != "" {
		conn = conn.Where("email = ?", reqUser.Email().GetValue())
	}

	if !reqUser.Name().GetIsNil() && reqUser.Name().GetValue() != "" {
		conn = conn.Where("full_name LIKE ?", "%"+reqUser.Name().GetValue()+"%")
	}

	result := conn.Find(&users)
	if result.Error != nil {
		err = result.Error
		return
	}

	userArgs := make([]groupObject.NewUserArgs, 0, len(users))
	for _, user := range users {
		userArgs = append(userArgs, groupObject.NewUserArgs{
			ID:    &user.ID,
			Name:  stringFromNullString(user.FullName),
			Email: &user.Email,
		})
	}

	// logger.Logging(
	// 	ctx,
	// 	middlewareRequestContext.GetRequestContext(ctx).TraceID.GetValue(),
	// )

	resUserList, err = groupObject.ReconstructUserList(&groupObject.NewUserListArgs{
		Content: userArgs,
	})
	if err != nil {
		logger.Logging(ctx, err)
		return
	}

	return
}
