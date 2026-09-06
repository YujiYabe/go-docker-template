package usecase

import (
	"context"

	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
)

// NewUseCase ...
func NewUseCase(
	toDomain ToDomain,
	toGatewayDB ToGatewayDB,
	toGatewayExternal ToGatewayExternal,
) (
	toUseCase ToUseCase,
) {
	toUseCase = &useCase{
		ToDomain:          toDomain,
		ToGatewayDB:       toGatewayDB,
		ToGatewayExternal: toGatewayExternal,
	}

	return
}

type (
	// useCase ...
	useCase struct {
		ToDomain          ToDomain
		ToGatewayDB       ToGatewayDB
		ToGatewayExternal ToGatewayExternal
	}
)

type (
	// ToDomain ...
	ToDomain interface {
		EnsurePrimaryEmploymentAssignable(
			user groupObject.User,
			userEmployment groupObject.UserEmployment,
		) (
			err error,
		)
	}
)

type (
	// ToGatewayDB ...
	ToGatewayDB interface {
		RunInTransaction(
			ctx context.Context,
			fn func(context.Context) error,
		) (
			err error,
		)

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

		UpdateUser(
			ctx context.Context,
			newUser groupObject.User,
		) (
			err error,
		)

		UpdateUserEmployment(
			ctx context.Context,
			userEmployment groupObject.UserEmployment,
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

	// ToGatewayExternal ...
	ToGatewayExternal interface {
		FetchAccessToken(
			ctx context.Context,
			credential groupObject.Credential,
		) (
			accessToken typeObject.AccessToken,
			err error,
		)

		GetUserViaGRPC(
			ctx context.Context,
			reqUser groupObject.User,
		) (
			resUserList groupObject.UserList,
			err error,
		)

		PublishTestTopic(
			ctx context.Context,
		) (
			err error,
		)
	}
)

type (

	// ToUseCase ...
	ToUseCase interface {
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
