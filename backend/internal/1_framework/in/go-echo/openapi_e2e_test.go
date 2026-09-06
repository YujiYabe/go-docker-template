package goEcho

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"backend/internal/1_framework/in/go-echo/openapi"
	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
)

type fakeController struct {
	getUserListByConditionCalled bool
	createUserCalled             bool
	publishTestTopicCalled       bool
	getValidationWordsCalled     bool
	addValidationWordCalled      bool
	updateValidationWordCalled   bool
	deleteValidationWordCalled   bool
	lastReqUser                  groupObject.User
	lastTargetType               string
	lastIsBlacklist              bool
	lastWord                     string
	lastOldWord                  string
	lastNewWord                  string
	userList                     groupObject.UserList
	createdUser                  groupObject.User
	createUserErr                error
	validationWords              []string
}

func newE2ETestEcho(
	t *testing.T,
	controller *fakeController,
) (
	echo *echo.Echo,
) {
	t.Helper()

	echoEcho := NewEcho()
	openapi.RegisterHandlers(
		echoEcho,
		&ServerInterfaceImpl{
			Controller: controller,
		},
	)

	echo = echoEcho

	return
}

func TestOpenAPIE2E_V1UsersGet(
	t *testing.T,
) {
	name := "Alice"
	email := "alice@example.com"
	userID := 1
	userList, err := groupObject.NewUserList(
		&groupObject.NewUserListArgs{
			Content: []groupObject.NewUserArgs{
				{
					ID:    &userID,
					Name:  &name,
					Email: &email,
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("failed to build user list: %v", err)
	}

	controller := &fakeController{
		userList: userList,
	}
	echoEcho := newE2ETestEcho(t, controller)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/v1/users?name=Alice&email=alice@example.com",
		nil,
	)
	req.Header.Set(echo.HeaderAccept, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !controller.getUserListByConditionCalled {
		t.Fatal("expected GetUserListByCondition to be called")
	}
	if got := controller.lastReqUser.Name().GetValue(); got != name {
		t.Fatalf("expected query name %q, got %q", name, got)
	}
	if got := controller.lastReqUser.Email().GetValue(); got != email {
		t.Fatalf("expected query email %q, got %q", email, got)
	}

	var body []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("expected 1 user, got %d", len(body))
	}
	if body[0].ID != userID || body[0].Name != name || body[0].Email != email {
		t.Fatalf("unexpected response body: %+v", body[0])
	}
}

func TestOpenAPIE2E_V1UsersPost(
	t *testing.T,
) {
	name := "Bob"
	email := "bob@example.com"
	userID := 5
	createdUser, err := groupObject.ReconstructUser(&groupObject.NewUserArgs{
		ID:    &userID,
		Name:  &name,
		Email: &email,
	})
	if err != nil {
		t.Fatalf("failed to build created user: %v", err)
	}

	controller := &fakeController{createdUser: *createdUser}
	echoEcho := newE2ETestEcho(t, controller)

	reqBody := bytes.NewBufferString(`{"name":"Bob","email":"bob@example.com"}`)
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/v1/users",
		reqBody,
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var body struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if !controller.createUserCalled {
		t.Fatal("expected CreateUser to be called")
	}
	if controller.lastReqUser.HasIdentity() {
		t.Fatal("new user should not have identity before persistence")
	}
	if body.ID != userID || body.Name != name || body.Email != email {
		t.Fatalf("unexpected response body: %+v", body)
	}
}

func TestOpenAPIE2E_V1UsersPostReturnsBadRequestOnCreateFailure(
	t *testing.T,
) {
	controller := &fakeController{createUserErr: errors.New("create failed")}
	echoEcho := newE2ETestEcho(t, controller)
	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/v1/users",
		bytes.NewBufferString(`{"name":"Bob","email":"bob@example.com"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
	if !controller.createUserCalled {
		t.Fatal("expected CreateUser to be called")
	}
}

func TestOpenAPIE2E_V1HealthGet(
	t *testing.T,
) {
	controller := &fakeController{}
	echoEcho := newE2ETestEcho(t, controller)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/v1/health",
		nil,
	)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "OK" {
		t.Fatalf("expected body %q, got %q", "OK", rec.Body.String())
	}
}

func TestOpenAPIE2E_V1ToPubsubGet(
	t *testing.T,
) {
	controller := &fakeController{}
	echoEcho := newE2ETestEcho(t, controller)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/v1/to-pubsub",
		nil,
	)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !controller.publishTestTopicCalled {
		t.Fatal("expected PublishTestTopic to be called")
	}
}

func TestOpenAPIE2E_V1ValidationWordRulesGet(
	t *testing.T,
) {
	controller := &fakeController{validationWords: []string{"root", "禁止語"}}
	echoEcho := newE2ETestEcho(t, controller)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/v1/validation-word-rules?targetType=name&isBlacklist=true",
		nil,
	)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if !controller.getValidationWordsCalled {
		t.Fatal("expected GetValidationWords to be called")
	}
	if controller.lastTargetType != "name" || !controller.lastIsBlacklist {
		t.Fatalf("unexpected params: %s %v", controller.lastTargetType, controller.lastIsBlacklist)
	}

	var body []openapi.ValidationWordRule
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if len(body) != 2 || body[0].Word != "root" || body[1].Word != "禁止語" {
		t.Fatalf("unexpected response body: %+v", body)
	}
}

func TestOpenAPIE2E_V1ValidationWordRulesPost(
	t *testing.T,
) {
	controller := &fakeController{}
	echoEcho := newE2ETestEcho(t, controller)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		"/v1/validation-word-rules",
		bytes.NewBufferString(`{"targetType":"name","isBlacklist":true,"word":"root"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	var body openapi.ValidationWordRule
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.TargetType != "name" || !body.IsBlacklist || body.Word != "root" {
		t.Fatalf("unexpected response body: %+v", body)
	}
	if !controller.addValidationWordCalled {
		t.Fatal("expected AddValidationWord to be called")
	}
	if controller.lastTargetType != "name" || !controller.lastIsBlacklist || controller.lastWord != "root" {
		t.Fatalf("unexpected request: %+v", controller)
	}
}

func TestOpenAPIE2E_V1ValidationWordRulesPut(
	t *testing.T,
) {
	controller := &fakeController{}
	echoEcho := newE2ETestEcho(t, controller)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodPut,
		"/v1/validation-word-rules",
		bytes.NewBufferString(`{"targetType":"name","isBlacklist":true,"oldWord":"root","newWord":"admin"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, rec.Code, rec.Body.String())
	}
	if !controller.updateValidationWordCalled {
		t.Fatal("expected UpdateValidationWord to be called")
	}
	if controller.lastOldWord != "root" || controller.lastNewWord != "admin" {
		t.Fatalf("unexpected request: %+v", controller)
	}
}

func TestOpenAPIE2E_V1ValidationWordRulesDelete(
	t *testing.T,
) {
	controller := &fakeController{}
	echoEcho := newE2ETestEcho(t, controller)

	req := httptest.NewRequestWithContext(
		context.Background(),
		http.MethodDelete,
		"/v1/validation-word-rules",
		bytes.NewBufferString(`{"targetType":"name","isBlacklist":true,"word":"root"}`),
	)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	echoEcho.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d: %s", http.StatusNoContent, rec.Code, rec.Body.String())
	}
	if !controller.deleteValidationWordCalled {
		t.Fatal("expected DeleteValidationWord to be called")
	}
	if controller.lastWord != "root" {
		t.Fatalf("unexpected request: %+v", controller)
	}
}

func (receiver *fakeController) GetUserList(
	_ context.Context,
) (
	userList groupObject.UserList,
	err error,
) {
	userList, err = groupObject.UserList{}, nil

	return
}

func (receiver *fakeController) GetUserListByCondition(
	_ context.Context,
	reqUser groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	receiver.getUserListByConditionCalled = true
	receiver.lastReqUser = reqUser

	resUserList, err = receiver.userList, nil

	return
}

func (receiver *fakeController) CreateUser(
	_ context.Context,
	newUser groupObject.User,
) (
	createdUser groupObject.User,
	err error,
) {
	receiver.createUserCalled = true
	receiver.lastReqUser = newUser
	createdUser, err = receiver.createdUser, receiver.createUserErr

	return
}

func (receiver *fakeController) FetchAccessToken(
	_ context.Context,
	_ groupObject.Credential,
) (
	accessToken typeObject.AccessToken,
	err error,
) {
	accessToken, err = typeObject.AccessToken{}, nil

	return
}

func (receiver *fakeController) GetUserListViaGRPC(
	_ context.Context,
	_ groupObject.User,
) (
	resUserList groupObject.UserList,
	err error,
) {
	resUserList, err = groupObject.UserList{}, nil

	return
}

func (receiver *fakeController) UpdateUser(
	_ context.Context,
	_ groupObject.User,
) (
	err error,
) {
	err = nil

	return
}

func (receiver *fakeController) UpdateUserProfileWithPrimaryEmployment(
	_ context.Context,
	_ groupObject.User,
	_ groupObject.UserEmployment,
) (
	err error,
) {
	err = nil

	return
}

func (receiver *fakeController) PublishTestTopic(
	_ context.Context,
) (
	err error,
) {
	receiver.publishTestTopicCalled = true

	err = nil

	return
}

func (receiver *fakeController) GetValidationWords(
	_ context.Context,
	targetType string,
	isBlacklist bool,
) (
	words []string,
	err error,
) {
	receiver.getValidationWordsCalled = true
	receiver.lastTargetType = targetType
	receiver.lastIsBlacklist = isBlacklist

	words, err = receiver.validationWords, nil

	return
}

func (receiver *fakeController) AddValidationWord(
	_ context.Context,
	targetType string,
	isBlacklist bool,
	word string,
) (
	err error,
) {
	receiver.addValidationWordCalled = true
	receiver.lastTargetType = targetType
	receiver.lastIsBlacklist = isBlacklist
	receiver.lastWord = word

	err = nil

	return
}

func (receiver *fakeController) UpdateValidationWord(
	_ context.Context,
	targetType string,
	isBlacklist bool,
	oldWord string,
	newWord string,
) (
	err error,
) {
	receiver.updateValidationWordCalled = true
	receiver.lastTargetType = targetType
	receiver.lastIsBlacklist = isBlacklist
	receiver.lastOldWord = oldWord
	receiver.lastNewWord = newWord

	err = nil

	return
}

func (receiver *fakeController) DeleteValidationWord(
	_ context.Context,
	targetType string,
	isBlacklist bool,
	word string,
) (
	err error,
) {
	receiver.deleteValidationWordCalled = true
	receiver.lastTargetType = targetType
	receiver.lastIsBlacklist = isBlacklist
	receiver.lastWord = word

	err = nil

	return
}
