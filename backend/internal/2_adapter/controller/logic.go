package controller

import (
	"context"

	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
)

func (receiver *controller) GetUserList(
	ctx context.Context,
) (
	userList groupObject.UserList,
	err error,
) {
	userList, err = receiver.UseCase.GetUserList(ctx)

	return
}

func (receiver *controller) GetUserListByCondition(
	ctx context.Context,
	reqUser groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	resUserList, err = receiver.UseCase.GetUserListByCondition(
		ctx,
		reqUser,
	)

	return
}

func (receiver *controller) CreateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	createdUser groupObject.User,
	err error,
) {
	createdUser, err = receiver.UseCase.CreateUser(ctx, newUser)

	return
}

func (receiver *controller) FetchAccessToken(
	ctx context.Context,
	credential groupObject.Credential,
) (
	accessToken typeObject.AccessToken,
	err error,
) {
	accessToken, err = receiver.UseCase.FetchAccessToken(
		ctx,
		credential,
	)

	return
}

func (receiver *controller) GetUserListViaGRPC(
	ctx context.Context,
	reqUser groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	resUserList, err = receiver.UseCase.GetUserListViaGRPC(
		ctx,
		reqUser,
	)

	return
}

func (receiver *controller) UpdateUser(
	ctx context.Context,
	newUser groupObject.User,
) (
	err error,
) {
	err = receiver.UseCase.UpdateUser(
		ctx,
		newUser,
	)

	return
}

func (receiver *controller) UpdateUserProfileWithPrimaryEmployment(
	ctx context.Context,
	newUser groupObject.User,
	userEmployment groupObject.UserEmployment,
) (
	err error,
) {
	err = receiver.UseCase.UpdateUserProfileWithPrimaryEmployment(
		ctx,
		newUser, userEmployment,
	)

	return
}

func (receiver *controller) PublishTestTopic(
	ctx context.Context,
) (
	err error,
) {
	err = receiver.UseCase.PublishTestTopic(ctx)

	return
}

func (receiver *controller) GetValidationWords(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
) (
	words []string,
	err error,
) {
	words, err = receiver.UseCase.GetValidationWords(
		ctx,
		targetType, isBlacklist,
	)

	return
}

func (receiver *controller) AddValidationWord(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
	word string,
) (
	err error,
) {
	err = receiver.UseCase.AddValidationWord(
		ctx,
		targetType,
		isBlacklist,
		word,
	)

	return
}

func (receiver *controller) UpdateValidationWord(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
	oldWord string,
	newWord string,
) (
	err error,
) {
	err = receiver.UseCase.UpdateValidationWord(
		ctx,
		targetType,
		isBlacklist,
		oldWord,
		newWord,
	)

	return
}

func (receiver *controller) DeleteValidationWord(
	ctx context.Context,
	targetType string,
	isBlacklist bool,
	word string,
) (
	err error,
) {
	err = receiver.UseCase.DeleteValidationWord(
		ctx,
		targetType,
		isBlacklist,
		word,
	)

	return
}
