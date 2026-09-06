package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	domain "backend/internal/4_domain"
	groupObject "backend/internal/4_domain/group_object"
)

func TestGetUserListWrapsGatewayError(
	t *testing.T,
) {
	t.Parallel()

	gatewayErr := errors.New("db unavailable")
	gatewayDB := &fakeGatewayDB{getListErr: gatewayErr}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})

	_, err := useCase.GetUserList(context.Background())
	if !errors.Is(err, gatewayErr) {
		t.Fatalf("expected wrapped gateway error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "GetUserList") {
		t.Fatalf("expected usecase name in error, got: %v", err)
	}
	if !gatewayDB.getListCalled {
		t.Fatal("gateway should be called")
	}
}

func TestGetUserListByConditionRequiresCondition(
	t *testing.T,
) {
	gatewayDB := &fakeGatewayDB{}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, nil, nil, nil)

	_, err := useCase.GetUserListByCondition(context.Background(), user)
	if err == nil {
		t.Fatal("expected condition error")
	}
	if !strings.Contains(err.Error(), "user search condition is required") {
		t.Fatalf("unexpected error: %v", err)
	}
	if gatewayDB.getByConditionCalled {
		t.Fatal("gateway should not be called when condition is empty")
	}
}

func TestGetUserListByConditionGatewayCases(
	t *testing.T,
) {
	t.Parallel()

	gatewayErr := errors.New("query failed")
	tests := []struct {
		name        string
		gatewayErr  error
		wantErr     error
		wantErrText string
	}{
		{
			name: "success",
		},
		{
			name:        "gateway error",
			gatewayErr:  gatewayErr,
			wantErr:     gatewayErr,
			wantErrText: "GetUserListByCondition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gatewayDB := &fakeGatewayDB{getByConditionErr: tt.gatewayErr}
			useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
			user := newTestUser(t, nil, stringPointer("alice"), nil)

			_, err := useCase.GetUserListByCondition(context.Background(), user)
			if tt.wantErr == nil && err != nil {
				t.Fatalf("expected success, got: %v", err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected wrapped gateway error, got: %v", err)
			}
			if tt.wantErrText != "" && !strings.Contains(err.Error(), tt.wantErrText) {
				t.Fatalf("expected error to contain %q, got: %v", tt.wantErrText, err)
			}
			if !gatewayDB.getByConditionCalled {
				t.Fatal("gateway should be called")
			}
		})
	}
}

func TestCreateUserValidatesAndRunsInTransaction(
	t *testing.T,
) {
	createdUser := newTestUser(t, intPointer(5), stringPointer("alice"), stringPointer("alice@example.com"))
	gatewayDB := &fakeGatewayDB{createdUser: createdUser}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	newUser := newTestUser(t, nil, stringPointer("alice"), stringPointer("alice@example.com"))

	result, err := useCase.CreateUser(context.Background(), newUser)
	if err != nil {
		t.Fatalf("expected create success, got: %v", err)
	}
	if result.ID().GetValue() != 5 {
		t.Fatalf("expected generated ID 5, got: %d", result.ID().GetValue())
	}
	if !gatewayDB.getValidationWordsCalled {
		t.Fatal("validation words should be fetched")
	}
	if !gatewayDB.runInTransactionCalled || !gatewayDB.createUserCalled {
		t.Fatal("create should run inside transaction")
	}
}

func TestCreateUserRejectsInvalidLifecycleAndBlacklistedName(
	t *testing.T,
) {
	tests := []struct {
		name            string
		user            groupObject.User
		validationWords []string
	}{
		{
			name: "persisted user",
			user: newTestUser(
				t,
				intPointer(1),
				stringPointer("alice"),
				stringPointer("alice@example.com"),
			),
		},
		{
			name: "blacklisted name",
			user: newTestUser(
				t,
				nil,
				stringPointer("root user"),
				stringPointer("alice@example.com"),
			),
			validationWords: []string{"root"},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			gatewayDB := &fakeGatewayDB{validationWords: testCase.validationWords}
			useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})

			_, err := useCase.CreateUser(context.Background(), testCase.user)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if gatewayDB.runInTransactionCalled || gatewayDB.createUserCalled {
				t.Fatal("gateway create should not run after validation failure")
			}
		})
	}
}

func TestCreateUserWrapsGatewayError(
	t *testing.T,
) {
	createErr := errors.New("insert failed")
	gatewayDB := &fakeGatewayDB{createUserErr: createErr}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	newUser := newTestUser(t, nil, stringPointer("alice"), stringPointer("alice@example.com"))

	_, err := useCase.CreateUser(context.Background(), newUser)
	if !errors.Is(err, createErr) {
		t.Fatalf("expected wrapped gateway error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "CreateUser") {
		t.Fatalf("expected usecase name in error, got: %v", err)
	}
	if !gatewayDB.runInTransactionCalled || !gatewayDB.createUserCalled {
		t.Fatal("gateway create should run inside transaction")
	}
}

func TestUpdateUserRunsInTransaction(
	t *testing.T,
) {
	gatewayDB := &fakeGatewayDB{}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, intPointer(1), stringPointer("name"), stringPointer("test@example.com"))

	if err := useCase.UpdateUser(context.Background(), user); err != nil {
		t.Fatalf("expected update success, got: %v", err)
	}
	if !gatewayDB.runInTransactionCalled {
		t.Fatal("transaction boundary was not used")
	}
	if !gatewayDB.updateUserCalled {
		t.Fatal("update gateway was not called")
	}
	if !gatewayDB.getValidationWordsCalled {
		t.Fatal("validation words should be fetched")
	}
}

func TestUpdateUserRequiresIdentity(
	t *testing.T,
) {
	gatewayDB := &fakeGatewayDB{}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, nil, stringPointer("name"), stringPointer("test@example.com"))

	err := useCase.UpdateUser(context.Background(), user)
	if err == nil {
		t.Fatal("expected identity error")
	}
	if !strings.Contains(err.Error(), "user identity is required") {
		t.Fatalf("unexpected error: %v", err)
	}
	if gatewayDB.runInTransactionCalled {
		t.Fatal("transaction should not run when user lifecycle state is invalid")
	}
	if gatewayDB.getValidationWordsCalled {
		t.Fatal("validation words should not be fetched when user lifecycle state is invalid")
	}
	if gatewayDB.updateUserCalled {
		t.Fatal("gateway should not be called when user lifecycle state is invalid")
	}
}

func TestUpdateUserRejectsBlacklistedName(
	t *testing.T,
) {
	gatewayDB := &fakeGatewayDB{validationWords: []string{"root"}}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, intPointer(1), stringPointer("root user"), stringPointer("test@example.com"))

	err := useCase.UpdateUser(context.Background(), user)
	if err == nil {
		t.Fatal("expected blacklist validation error")
	}
	if !strings.Contains(err.Error(), "detect target spell") {
		t.Fatalf("unexpected error: %v", err)
	}
	if gatewayDB.runInTransactionCalled {
		t.Fatal("transaction should not run when name is blacklisted")
	}
	if gatewayDB.updateUserCalled {
		t.Fatal("gateway should not be called when name is blacklisted")
	}
}

func TestUpdateUserWrapsValidationWordFetchError(
	t *testing.T,
) {
	validationErr := errors.New("validation word query failed")
	gatewayDB := &fakeGatewayDB{getValidationWordsErr: validationErr}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, intPointer(1), stringPointer("name"), stringPointer("test@example.com"))

	err := useCase.UpdateUser(context.Background(), user)
	if !errors.Is(err, validationErr) {
		t.Fatalf("expected wrapped validation word error, got: %v", err)
	}
	if gatewayDB.runInTransactionCalled {
		t.Fatal("transaction should not run when validation words cannot be fetched")
	}
}

func TestUpdateUserWrapsGatewayError(
	t *testing.T,
) {
	t.Parallel()

	gatewayErr := errors.New("update failed")
	gatewayDB := &fakeGatewayDB{updateUserErr: gatewayErr}
	useCase := NewUseCase(nil, gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, intPointer(1), stringPointer("name"), stringPointer("test@example.com"))

	err := useCase.UpdateUser(context.Background(), user)
	if !errors.Is(err, gatewayErr) {
		t.Fatalf("expected wrapped gateway error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "UpdateUser") {
		t.Fatalf("expected usecase name in error, got: %v", err)
	}
	if !gatewayDB.runInTransactionCalled {
		t.Fatal("transaction boundary should be used")
	}
	if !gatewayDB.updateUserCalled {
		t.Fatal("gateway should be called")
	}
}

func TestUpdateUserProfileWithPrimaryEmploymentRunsMultipleUpdatesInTransaction(
	t *testing.T,
) {
	gatewayDB := &fakeGatewayDB{}
	useCase := NewUseCase(domain.NewDomain(), gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, intPointer(1), stringPointer("name"), stringPointer("test@example.com"))
	employment := newTestUserEmployment(t, 1, true)

	if err := useCase.UpdateUserProfileWithPrimaryEmployment(context.Background(), user, employment); err != nil {
		t.Fatalf("expected update success, got: %v", err)
	}
	if !gatewayDB.runInTransactionCalled {
		t.Fatal("transaction boundary was not used")
	}
	if !gatewayDB.updateUserCalled {
		t.Fatal("user update gateway was not called")
	}
	if !gatewayDB.updateEmploymentCalled {
		t.Fatal("employment update gateway was not called")
	}
	if !gatewayDB.getValidationWordsCalled {
		t.Fatal("validation words should be fetched")
	}
	if strings.Join(gatewayDB.calls, ",") != "update_user,update_user_employment" {
		t.Fatalf("unexpected call order: %v", gatewayDB.calls)
	}
}

func TestUpdateUserProfileWithPrimaryEmploymentRejectsDifferentUser(
	t *testing.T,
) {
	gatewayDB := &fakeGatewayDB{}
	useCase := NewUseCase(domain.NewDomain(), gatewayDB, &fakeGatewayExternal{})
	user := newTestUser(t, intPointer(1), stringPointer("name"), stringPointer("test@example.com"))
	employment := newTestUserEmployment(t, 2, true)

	err := useCase.UpdateUserProfileWithPrimaryEmployment(context.Background(), user, employment)
	if err == nil {
		t.Fatal("expected ownership error")
	}
	if !strings.Contains(err.Error(), "must belong to the user") {
		t.Fatalf("unexpected error: %v", err)
	}
	if gatewayDB.runInTransactionCalled {
		t.Fatal("transaction should not run when employment does not belong to user")
	}
	if gatewayDB.updateUserCalled || gatewayDB.updateEmploymentCalled {
		t.Fatal("gateway should not be called when employment does not belong to user")
	}
}

func TestUpdateUserProfileWithPrimaryEmploymentGatewayErrors(
	t *testing.T,
) {
	t.Parallel()

	updateUserErr := errors.New("user update failed")
	updateEmploymentErr := errors.New("employment update failed")
	tests := []struct {
		name                    string
		updateUserErr           error
		updateUserEmploymentErr error
		wantErr                 error
		wantCalls               string
	}{
		{
			name:          "user update fails before employment update",
			updateUserErr: updateUserErr,
			wantErr:       updateUserErr,
			wantCalls:     "update_user",
		},
		{
			name:                    "employment update fails after user update",
			updateUserEmploymentErr: updateEmploymentErr,
			wantErr:                 updateEmploymentErr,
			wantCalls:               "update_user,update_user_employment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gatewayDB := &fakeGatewayDB{
				updateUserErr:           tt.updateUserErr,
				updateUserEmploymentErr: tt.updateUserEmploymentErr,
			}
			useCase := NewUseCase(domain.NewDomain(), gatewayDB, &fakeGatewayExternal{})
			user := newTestUser(t, intPointer(1), stringPointer("name"), stringPointer("test@example.com"))
			employment := newTestUserEmployment(t, 1, true)

			err := useCase.UpdateUserProfileWithPrimaryEmployment(context.Background(), user, employment)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected wrapped gateway error, got: %v", err)
			}
			if !strings.Contains(err.Error(), "UpdateUserProfileWithPrimaryEmployment") {
				t.Fatalf("expected usecase name in error, got: %v", err)
			}
			if !gatewayDB.runInTransactionCalled {
				t.Fatal("transaction boundary should be used")
			}
			if strings.Join(gatewayDB.calls, ",") != tt.wantCalls {
				t.Fatalf("unexpected call order: %v", gatewayDB.calls)
			}
		})
	}
}
