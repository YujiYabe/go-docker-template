package usecase

import (
	"context"
	"fmt"

	groupObject "backend/internal/4_domain/group_object"
)

const validationWordRuleTargetName = "name"

func (receiver *useCase) GetUserList(
	ctx context.Context,
) (
	userList groupObject.UserList,
	err error,
) {
	userList = groupObject.UserList{}
	if err = ensureContextReady(ctx, "GetUserList"); err != nil {

		return
	}
	userList, err = receiver.ToGatewayDB.GetUserList(ctx)
	if err != nil {
		err = fmt.Errorf("GetUserList: %w", err)
	}

	return
}

func (receiver *useCase) GetUserListByCondition(
	ctx context.Context,
	reqUser groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	resUserList = groupObject.UserList{}
	if err = ensureContextReady(ctx, "GetUserListByCondition"); err != nil {

		return
	}
	if !reqUser.CanBeUsedAsSearchCondition() {
		err = fmt.Errorf("GetUserListByCondition: user search condition is required")

		return
	}

	resUserList, err = receiver.ToGatewayDB.GetUserListByCondition(
		ctx,
		reqUser,
	)
	if err != nil {
		err = fmt.Errorf("GetUserListByCondition: %w", err)
	}

	return
}

func (receiver *useCase) CreateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	createdUser groupObject.User,
	err error,
) {
	if err = ensureContextReady(ctx, "CreateUser"); err != nil {
		return
	}
	if err = newUser.EnsureReadyToCreate(); err != nil {
		err = fmt.Errorf("CreateUser: %w", err)

		return
	}

	nameBlacklist, err := receiver.ToGatewayDB.GetValidationWords(
		ctx,
		validationWordRuleTargetName,
		true,
	)
	if err != nil {
		err = fmt.Errorf("CreateUser: %w", err)

		return
	}
	if err = newUser.ValidateNameBlacklist(nameBlacklist); err != nil {
		err = fmt.Errorf("CreateUser: %w", err)

		return
	}

	err = receiver.ToGatewayDB.RunInTransaction(
		ctx,
		func(txCtx context.Context) (transactionErr error) {
			createdUser, transactionErr = receiver.ToGatewayDB.CreateUser(txCtx, newUser)

			return
		},
	)
	if err != nil {
		err = fmt.Errorf("CreateUser: %w", err)
	}

	return
}

func (receiver *useCase) UpdateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	err error,
) {
	if returnedErr := ensureContextReady(ctx, "UpdateUser"); returnedErr != nil {
		err = returnedErr
		return
	}
	if err := newUser.EnsureReadyToUpdate(); err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}
	nameBlacklist, err := receiver.ToGatewayDB.GetValidationWords(
		ctx,
		validationWordRuleTargetName,
		true,
	)
	if err != nil {
		err = fmt.Errorf("UpdateUser: %w", err)
		return
	}
	if err := newUser.ValidateNameBlacklist(nameBlacklist); err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}

	if err := receiver.ToGatewayDB.RunInTransaction(
		ctx,
		func(txCtx context.Context) error {
			return receiver.ToGatewayDB.UpdateUser(txCtx, newUser)
		},
	); err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}

	err = nil
	return
}

func (receiver *useCase) GetUserListViaGRPC(
	ctx context.Context,
	reqUser groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	resUserList = groupObject.UserList{}
	if err = ensureContextReady(ctx, "GetUserListViaGRPC"); err != nil {

		return
	}
	if !reqUser.CanBeUsedAsSearchCondition() {
		err = fmt.Errorf("GetUserListViaGRPC: user search condition is required")

		return
	}
	resUserList, err = receiver.ToGatewayExternal.GetUserViaGRPC(
		ctx,
		reqUser,
	)
	if err != nil {
		err = fmt.Errorf("GetUserListViaGRPC: %w", err)
	}

	return
}

func (receiver *useCase) UpdateUserProfileWithPrimaryEmployment(
	ctx context.Context,
	newUser groupObject.User,
	userEmployment groupObject.UserEmployment,
) (
	err error,
) {
	if returnedErr := ensureContextReady(ctx, "UpdateUserProfileWithPrimaryEmployment"); returnedErr != nil {
		err = returnedErr
		return
	}
	if err := newUser.EnsureReadyToUpdate(); err != nil {
		return fmt.Errorf("UpdateUserProfileWithPrimaryEmployment: %w", err)
	}
	nameBlacklist, err := receiver.ToGatewayDB.GetValidationWords(
		ctx,
		validationWordRuleTargetName,
		true,
	)
	if err != nil {
		err = fmt.Errorf("UpdateUserProfileWithPrimaryEmployment: %w", err)
		return
	}
	if err := newUser.ValidateNameBlacklist(nameBlacklist); err != nil {
		return fmt.Errorf("UpdateUserProfileWithPrimaryEmployment: %w", err)
	}
	if err := receiver.ToDomain.EnsurePrimaryEmploymentAssignable(newUser, userEmployment); err != nil {
		return fmt.Errorf("UpdateUserProfileWithPrimaryEmployment: %w", err)
	}

	if err := receiver.ToGatewayDB.RunInTransaction(
		ctx,
		func(txCtx context.Context) error {
			if err := receiver.ToGatewayDB.UpdateUser(txCtx, newUser); err != nil {
				return err
			}
			return receiver.ToGatewayDB.UpdateUserEmployment(txCtx, userEmployment)
		},
	); err != nil {
		return fmt.Errorf("UpdateUserProfileWithPrimaryEmployment: %w", err)
	}

	err = nil
	return
}
