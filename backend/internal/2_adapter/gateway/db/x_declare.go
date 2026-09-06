package db_gateway

import (
	"context"

	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
)

type GatewayDB struct {
	ToPostgres ToPostgres
	ToRedis    ToRedis
}

// NewGatewayDB ...
func NewGatewayDB(
	toPostgres ToPostgres,
	toRedis ToRedis,
) (
	gatewayDB *GatewayDB,
) {
	gatewayDB = &GatewayDB{
		ToPostgres: toPostgres,
		ToRedis:    toRedis,
	}

	return
}

type (

	// ToPostgres ...
	ToPostgres interface {
		RunInTransaction(
			ctx context.Context,
			fn func(context.Context) error,
		) (
			err error,
		)

		GetUser(
			ctx context.Context,
			id typeObject.ID,
		) (
			user groupObject.User,
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

	// ToRedis ...
	ToRedis interface {
		ResetPlaceListInRedis(
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
			hit bool,
			err error,
		)

		SetValidationWords(
			ctx context.Context,
			targetType string,
			isBlacklist bool,
			words []string,
		) (
			err error,
		)

		DeleteValidationWordsCache(
			ctx context.Context,
			targetType string,
			isBlacklist bool,
		) (
			err error,
		)
	}
)
