package controller

import (
	"context"

	usecase "backend/internal/3_usecase"

	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
)

// NewController ...
func NewController(
	useCase usecase.ToUseCase,
) (
	toController ToController,
) {
	toController = &controller{
		UseCase: useCase,
	}

	return
}

type (
	// controller ...
	controller struct {
		UseCase usecase.ToUseCase
	}

	// ToController ...
	ToController interface {
		GetUserList(
			ctx context.Context,
		) (
			userList groupObject.UserList,
			err error,
		)

		GetUserListByCondition(
			ctx context.Context,
			reqUser groupObject.User,
		) (
			resUserList groupObject.UserList,
			err error,
		)

		CreateUser(
			ctx context.Context,
			newUser groupObject.User,
		) (
			createdUser groupObject.User,
			err error,
		)

		FetchAccessToken(
			ctx context.Context,
			credential groupObject.Credential,
		) (
			accessToken typeObject.AccessToken,
			err error,
		)

		GetUserListViaGRPC(
			ctx context.Context,
			reqUser groupObject.User,
		) (
			resUserList groupObject.UserList,
			err error,
		)

		UpdateUser(
			ctx context.Context,
			newUser groupObject.User,
		) (
			err error,
		)

		UpdateUserProfileWithPrimaryEmployment(
			ctx context.Context,
			newUser groupObject.User,
			userEmployment groupObject.UserEmployment,
		) (
			err error,
		)

		PublishTestTopic(
			ctx context.Context,
		) (
			err error,
		)

		GetValidationWords(
			ctx context.Context,
			targetType string,
			isBlacklist bool,
		) (
			words []string,
			err error,
		)

		AddValidationWord(
			ctx context.Context,
			targetType string,
			isBlacklist bool,
			word string,
		) (
			err error,
		)

		UpdateValidationWord(
			ctx context.Context,
			targetType string,
			isBlacklist bool,
			oldWord string,
			newWord string,
		) (
			err error,
		)

		DeleteValidationWord(
			ctx context.Context,
			targetType string,
			isBlacklist bool,
			word string,
		) (
			err error,
		)
	}
)
