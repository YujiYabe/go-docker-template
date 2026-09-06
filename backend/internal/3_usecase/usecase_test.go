package usecase

import (
	"context"
	"errors"
	"testing"

	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
)

type fakeGatewayDB struct {
	runInTransactionCalled     bool
	createUserCalled           bool
	updateUserCalled           bool
	updateEmploymentCalled     bool
	getListCalled              bool
	getByConditionCalled       bool
	getValidationWordsCalled   bool
	addValidationWordCalled    bool
	updateValidationWordCalled bool
	deleteValidationWordCalled bool
	calls                      []string
	getListErr                 error
	getByConditionErr          error
	createUserErr              error
	updateUserErr              error
	updateUserEmploymentErr    error
	getValidationWordsErr      error
	validationWordUpdateErr    error
	validationWords            []string
	createdUser                groupObject.User
}

func (receiver *fakeGatewayDB) CreateUser(
	_ context.Context,
	_ groupObject.User,
) (
	createdUser groupObject.User,
	err error,
) {
	receiver.createUserCalled = true
	receiver.calls = append(receiver.calls, "create_user")
	createdUser, err = receiver.createdUser, receiver.createUserErr

	return
}

func (receiver *fakeGatewayDB) RunInTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) (
	err error,
) {
	receiver.runInTransactionCalled = true
	err = fn(ctx)

	return
}

func (receiver *fakeGatewayDB) GetUserList(
	_ context.Context,
) (
	userList groupObject.UserList,
	err error,
) {
	receiver.getListCalled = true
	userList, err = groupObject.UserList{}, receiver.getListErr

	return
}

func (receiver *fakeGatewayDB) GetUserListByCondition(
	_ context.Context,
	_ groupObject.User,
) (
	userList groupObject.UserList,
	err error,
) {
	receiver.getByConditionCalled = true
	userList, err = groupObject.UserList{}, receiver.getByConditionErr

	return
}

func (receiver *fakeGatewayDB) UpdateUser(
	_ context.Context,
	_ groupObject.User,
) (
	err error,
) {
	receiver.updateUserCalled = true
	receiver.calls = append(receiver.calls, "update_user")
	err = receiver.updateUserErr

	return
}

func (receiver *fakeGatewayDB) UpdateUserEmployment(
	_ context.Context,
	_ groupObject.UserEmployment,
) (
	err error,
) {
	receiver.updateEmploymentCalled = true
	receiver.calls = append(receiver.calls, "update_user_employment")
	err = receiver.updateUserEmploymentErr

	return
}

func (receiver *fakeGatewayDB) GetValidationWords(
	_ context.Context,
	_ string,
	_ bool,
) (
	words []string,
	err error,
) {
	receiver.getValidationWordsCalled = true
	words, err = receiver.validationWords, receiver.getValidationWordsErr

	return
}

func (receiver *fakeGatewayDB) AddValidationWord(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
) (
	err error,
) {
	receiver.addValidationWordCalled = true
	err = receiver.validationWordUpdateErr

	return
}

func (receiver *fakeGatewayDB) UpdateValidationWord(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
	_ string,
) (
	err error,
) {
	receiver.updateValidationWordCalled = true
	err = receiver.validationWordUpdateErr

	return
}

func (receiver *fakeGatewayDB) DeleteValidationWord(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
) (
	err error,
) {
	receiver.deleteValidationWordCalled = true
	err = receiver.validationWordUpdateErr

	return
}

type fakeGatewayExternal struct {
	fetchAccessTokenCalled bool
	viaGRPCCalled          bool
	publishTestTopicCalled bool
	fetchAccessTokenErr    error
	viaGRPCErr             error
	publishTestTopicErr    error
}

func (receiver *fakeGatewayExternal) FetchAccessToken(
	_ context.Context,
	_ groupObject.Credential,
) (
	accessToken typeObject.AccessToken,
	err error,
) {
	receiver.fetchAccessTokenCalled = true
	if receiver.fetchAccessTokenErr != nil {
		accessToken, err = typeObject.AccessToken{}, receiver.fetchAccessTokenErr

		return
	}
	accessToken, err = typeObject.NewAccessToken(stringPointer("access-token"))

	return
}

func (receiver *fakeGatewayExternal) GetUserViaGRPC(
	_ context.Context,
	_ groupObject.User,
) (
	userList groupObject.UserList,
	err error,
) {
	receiver.viaGRPCCalled = true
	userList, err = groupObject.UserList{}, receiver.viaGRPCErr

	return
}

func (receiver *fakeGatewayExternal) PublishTestTopic(
	_ context.Context,
) (
	err error,
) {
	receiver.publishTestTopicCalled = true
	err = receiver.publishTestTopicErr

	return
}

func TestEnsureContextReadyReturnsCanceledError(
	t *testing.T,
) {
	useCase := NewUseCase(nil, &fakeGatewayDB{}, &fakeGatewayExternal{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := useCase.GetUserList(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func newTestUser(
	t *testing.T,
	id *int,
	name *string,
	email *string,
) (
	user groupObject.User,
) {
	t.Helper()

	newUser, err := groupObject.NewUser(&groupObject.NewUserArgs{
		ID:    id,
		Name:  name,
		Email: email,
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	user = *newUser

	return
}

func newTestCredential(
	t *testing.T,
	clientID string,
	clientSecret string,
) (
	credential groupObject.Credential,
) {
	t.Helper()

	newCredential, err := groupObject.NewCredential(&groupObject.NewCredentialArgs{
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
	})
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	credential = *newCredential

	return
}

func newTestUserEmployment(
	t *testing.T,
	userID int,
	isPrimary bool,
) (
	userEmployment groupObject.UserEmployment,
) {
	t.Helper()

	employeeCode := "EMP001"
	employmentType := "full_time"
	newUserEmployment, err := groupObject.NewUserEmployment(&groupObject.NewUserEmploymentArgs{
		UserID:         intPointer(userID),
		CompanyID:      intPointer(1),
		DepartmentID:   intPointer(2),
		PositionID:     intPointer(3),
		EmployeeCode:   &employeeCode,
		EmploymentType: &employmentType,
		IsPrimary:      &isPrimary,
	})
	if err != nil {
		t.Fatalf("failed to create user employment: %v", err)
	}

	userEmployment = *newUserEmployment

	return
}

func intPointer(
	value int,
) (
	valuePointer *int,
) {
	valuePointer = &value

	return
}

func stringPointer(
	value string,
) (
	valuePointer *string,
) {
	valuePointer = &value

	return
}
