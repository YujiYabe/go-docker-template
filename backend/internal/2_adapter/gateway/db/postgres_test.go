package db_gateway

import (
	"context"
	"errors"
	"reflect"
	"testing"

	groupObject "backend/internal/4_domain/group_object"
	typeObject "backend/internal/4_domain/type_object"
)

func TestGatewayDBGetValidationWordsReturnsRedisCacheHit(
	t *testing.T,
) {
	postgres := &fakePostgres{}
	redis := &fakeRedis{
		words: []string{"root"},
		hit:   true,
	}
	gateway := NewGatewayDB(postgres, redis)

	words, err := gateway.GetValidationWords(context.Background(), "name", true)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !reflect.DeepEqual(words, []string{"root"}) {
		t.Fatalf("unexpected words: %v", words)
	}
	if postgres.getValidationWordsCalled {
		t.Fatal("postgres should not be called on cache hit")
	}
	if redis.setValidationWordsCalled {
		t.Fatal("cache should not be written on cache hit")
	}
}

func TestGatewayDBGetValidationWordsFillsRedisOnCacheMiss(
	t *testing.T,
) {
	postgres := &fakePostgres{words: []string{"admin", "root"}}
	redis := &fakeRedis{}
	gateway := NewGatewayDB(postgres, redis)

	words, err := gateway.GetValidationWords(context.Background(), "name", true)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !reflect.DeepEqual(words, []string{"admin", "root"}) {
		t.Fatalf("unexpected words: %v", words)
	}
	if !postgres.getValidationWordsCalled {
		t.Fatal("postgres should be called on cache miss")
	}
	if !redis.setValidationWordsCalled {
		t.Fatal("cache should be written after postgres success")
	}
}

func TestGatewayDBGetValidationWordsFallsBackWhenRedisFails(
	t *testing.T,
) {
	redisErr := errors.New("redis unavailable")
	postgres := &fakePostgres{words: []string{"root"}}
	redis := &fakeRedis{getErr: redisErr}
	gateway := NewGatewayDB(postgres, redis)

	words, err := gateway.GetValidationWords(context.Background(), "name", true)
	if err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !reflect.DeepEqual(words, []string{"root"}) {
		t.Fatalf("unexpected words: %v", words)
	}
	if !postgres.getValidationWordsCalled {
		t.Fatal("postgres should be called when redis fails")
	}
}

func TestGatewayDBGetValidationWordsReturnsPostgresError(
	t *testing.T,
) {
	postgresErr := errors.New("postgres unavailable")
	postgres := &fakePostgres{err: postgresErr}
	redis := &fakeRedis{}
	gateway := NewGatewayDB(postgres, redis)

	_, err := gateway.GetValidationWords(context.Background(), "name", true)
	if !errors.Is(err, postgresErr) {
		t.Fatalf("expected postgres error, got: %v", err)
	}
	if redis.setValidationWordsCalled {
		t.Fatal("cache should not be written after postgres error")
	}
}

func TestGatewayDBAddValidationWordDeletesCacheAfterPostgresSuccess(
	t *testing.T,
) {
	postgres := &fakePostgres{}
	redis := &fakeRedis{}
	gateway := NewGatewayDB(postgres, redis)

	if err := gateway.AddValidationWord(context.Background(), "name", true, "root"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !postgres.addValidationWordCalled {
		t.Fatal("postgres add should be called")
	}
	if !redis.deleteValidationWordsCacheCalled {
		t.Fatal("cache should be deleted after postgres success")
	}
}

func TestGatewayDBUpdateValidationWordDeletesCacheAfterPostgresSuccess(
	t *testing.T,
) {
	postgres := &fakePostgres{}
	redis := &fakeRedis{}
	gateway := NewGatewayDB(postgres, redis)

	if err := gateway.UpdateValidationWord(context.Background(), "name", true, "root", "admin"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !postgres.updateValidationWordCalled {
		t.Fatal("postgres update should be called")
	}
	if !redis.deleteValidationWordsCacheCalled {
		t.Fatal("cache should be deleted after postgres success")
	}
}

func TestGatewayDBDeleteValidationWordDeletesCacheAfterPostgresSuccess(
	t *testing.T,
) {
	postgres := &fakePostgres{}
	redis := &fakeRedis{}
	gateway := NewGatewayDB(postgres, redis)

	if err := gateway.DeleteValidationWord(context.Background(), "name", true, "root"); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if !postgres.deleteValidationWordCalled {
		t.Fatal("postgres delete should be called")
	}
	if !redis.deleteValidationWordsCacheCalled {
		t.Fatal("cache should be deleted after postgres success")
	}
}

func TestGatewayDBValidationWordUpdateDoesNotDeleteCacheAfterPostgresError(
	t *testing.T,
) {
	postgresErr := errors.New("postgres update failed")
	postgres := &fakePostgres{err: postgresErr}
	redis := &fakeRedis{}
	gateway := NewGatewayDB(postgres, redis)

	err := gateway.AddValidationWord(context.Background(), "name", true, "root")
	if !errors.Is(err, postgresErr) {
		t.Fatalf("expected postgres error, got: %v", err)
	}
	if redis.deleteValidationWordsCacheCalled {
		t.Fatal("cache should not be deleted after postgres error")
	}
}

func TestGatewayDBValidationWordUpdateIgnoresCacheDeleteError(
	t *testing.T,
) {
	postgres := &fakePostgres{}
	redis := &fakeRedis{deleteErr: errors.New("redis delete failed")}
	gateway := NewGatewayDB(postgres, redis)

	if err := gateway.AddValidationWord(context.Background(), "name", true, "root"); err != nil {
		t.Fatalf("cache delete error should not fail update, got: %v", err)
	}
	if !redis.deleteValidationWordsCacheCalled {
		t.Fatal("cache delete should be attempted")
	}
}

type fakePostgres struct {
	getValidationWordsCalled   bool
	addValidationWordCalled    bool
	updateValidationWordCalled bool
	deleteValidationWordCalled bool
	words                      []string
	err                        error
}

func (receiver *fakePostgres) RunInTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) (
	err error,
) {
	err = fn(ctx)

	return
}

func (receiver *fakePostgres) GetUser(
	_ context.Context,
	_ typeObject.ID,
) (
	user groupObject.User,
	err error,
) {
	user, err = groupObject.User{}, nil

	return
}

func (receiver *fakePostgres) GetUserList(
	_ context.Context,
) (
	userList groupObject.UserList,
	err error,
) {
	userList, err = groupObject.UserList{}, nil

	return
}

func (receiver *fakePostgres) GetUserListByCondition(
	_ context.Context,
	_ groupObject.User,
) (
	userList groupObject.UserList,
	err error,
) {
	userList, err = groupObject.UserList{}, nil

	return
}

func (receiver *fakePostgres) CreateUser(
	_ context.Context,
	_ groupObject.User,
) (
	createdUser groupObject.User,
	err error,
) {
	createdUser, err = groupObject.User{}, receiver.err

	return
}

func (receiver *fakePostgres) UpdateUser(
	_ context.Context,
	_ groupObject.User,
) (
	err error,
) {
	err = nil

	return
}

func (receiver *fakePostgres) UpdateUserEmployment(
	_ context.Context,
	_ groupObject.UserEmployment,
) (
	err error,
) {
	err = nil

	return
}

func (receiver *fakePostgres) GetValidationWords(
	_ context.Context,
	_ string,
	_ bool,
) (
	words []string,
	err error,
) {
	receiver.getValidationWordsCalled = true
	words, err = receiver.words, receiver.err

	return
}

func (receiver *fakePostgres) AddValidationWord(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
) (
	err error,
) {
	receiver.addValidationWordCalled = true
	err = receiver.err

	return
}

func (receiver *fakePostgres) UpdateValidationWord(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
	_ string,
) (
	err error,
) {
	receiver.updateValidationWordCalled = true
	err = receiver.err

	return
}

func (receiver *fakePostgres) DeleteValidationWord(
	_ context.Context,
	_ string,
	_ bool,
	_ string,
) (
	err error,
) {
	receiver.deleteValidationWordCalled = true
	err = receiver.err

	return
}

type fakeRedis struct {
	setValidationWordsCalled         bool
	deleteValidationWordsCacheCalled bool
	words                            []string
	hit                              bool
	getErr                           error
	setErr                           error
	deleteErr                        error
}

func (receiver *fakeRedis) ResetPlaceListInRedis(
	_ context.Context,
) (
	err error,
) {
	err = nil

	return
}

func (receiver *fakeRedis) GetValidationWords(
	_ context.Context,
	_ string,
	_ bool,
) (
	words []string,
	ok bool,
	err error,
) {
	words, ok, err = receiver.words, receiver.hit, receiver.getErr

	return
}

func (receiver *fakeRedis) SetValidationWords(
	_ context.Context,
	_ string,
	_ bool,
	_ []string,
) (
	err error,
) {
	receiver.setValidationWordsCalled = true
	err = receiver.setErr

	return
}

func (receiver *fakeRedis) DeleteValidationWordsCache(
	_ context.Context,
	_ string,
	_ bool,
) (
	err error,
) {
	receiver.deleteValidationWordsCacheCalled = true
	err = receiver.deleteErr

	return
}
