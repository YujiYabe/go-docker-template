package db_gateway

import (
	"context"

	groupObject "backend/internal/4_domain/group_object"
	"backend/internal/logger"
)

func (receiver *GatewayDB) RunInTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) (
	err error,
) {
	err = receiver.ToPostgres.RunInTransaction(ctx, fn)

	return
}

// GetUserList ...
func (receiver *GatewayDB) GetUserList(
	ctx context.Context,
) (
	userList groupObject.UserList,
	err error,
) {
	userList, err = receiver.ToPostgres.GetUserList(
		ctx,
	)

	return
}

// GetUserListByCondition ...
func (receiver *GatewayDB) GetUserListByCondition(
	ctx context.Context,
	reqUser groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	resUserList, err = receiver.ToPostgres.GetUserListByCondition(
		ctx,
		reqUser,
	)

	return
}

func (receiver *GatewayDB) CreateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	createdUser groupObject.User,
	err error,
) {
	createdUser, err = receiver.ToPostgres.CreateUser(ctx, newUser)

	return
}

// UpdateUser ...
func (receiver *GatewayDB) UpdateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	err error,
) {
	err = receiver.ToPostgres.UpdateUser(ctx, newUser)

	return
}

func (receiver *GatewayDB) UpdateUserEmployment(
	ctx context.Context,
	userEmployment groupObject.UserEmployment,
) (
	err error,
) {
	err = receiver.ToPostgres.UpdateUserEmployment(ctx, userEmployment)

	return
}

func (receiver *GatewayDB) GetValidationWords(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
) (
	words []string,
	err error,
) {
	if receiver.ToRedis != nil {
		words, hit, redisErr := receiver.ToRedis.GetValidationWords(ctx, targetType, isBlacklist)
		if redisErr != nil {
			logger.Logging(ctx, redisErr)
		}
		if redisErr == nil && hit {
			return words, nil
		}
	}

	words, err = receiver.ToPostgres.GetValidationWords(ctx, targetType, isBlacklist)
	if err != nil {
		words = nil
		return
	}

	if receiver.ToRedis != nil {
		if redisErr := receiver.ToRedis.SetValidationWords(ctx, targetType, isBlacklist, words); redisErr != nil {
			logger.Logging(ctx, redisErr)
		}
	}

	err = nil

	return
}

func (receiver *GatewayDB) AddValidationWord(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
	word string,
) (
	err error,
) {
	if returnedErr := receiver.ToPostgres.AddValidationWord(ctx, targetType, isBlacklist, word); returnedErr != nil {
		err = returnedErr

		return
	}

	receiver.deleteValidationWordsCache(ctx, targetType, isBlacklist)
	err = nil

	return
}

func (receiver *GatewayDB) UpdateValidationWord(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
	oldWord string,
	newWord string,
) (
	err error,
) {
	if returnedErr := receiver.ToPostgres.UpdateValidationWord(ctx, targetType, isBlacklist, oldWord, newWord); returnedErr != nil {
		err = returnedErr

		return
	}

	receiver.deleteValidationWordsCache(ctx, targetType, isBlacklist)
	err = nil

	return
}

func (receiver *GatewayDB) DeleteValidationWord(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
	word string,
) (
	err error,
) {
	if returnedErr := receiver.ToPostgres.DeleteValidationWord(ctx, targetType, isBlacklist, word); returnedErr != nil {
		err = returnedErr

		return
	}

	receiver.deleteValidationWordsCache(ctx, targetType, isBlacklist)
	err = nil

	return
}

func (receiver *GatewayDB) deleteValidationWordsCache(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
) {
	if receiver.ToRedis == nil {

		return
	}

	if err := receiver.ToRedis.DeleteValidationWordsCache(ctx, targetType, isBlacklist); err != nil {
		logger.Logging(ctx, err)
	}
}
