package group_object

import "testing"

func TestNewUserAllowsDraftWithoutIdentity(
	t *testing.T,
) {
	user, err := NewUser(&NewUserArgs{
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err != nil {
		t.Fatalf("expected draft user, got: %v", err)
	}

	if user.HasIdentity() {
		t.Fatal("draft user should not have identity")
	}
	if !user.CanBeUsedAsSearchCondition() {
		t.Fatal("user with name should be usable as search condition")
	}
}

func TestReconstructUserRequiresIdentity(
	t *testing.T,
) {
	_, err := ReconstructUser(&NewUserArgs{
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err == nil {
		t.Fatal("expected identity error")
	}
}

func TestNewUserSearchConditionRequiresCondition(
	t *testing.T,
) {
	_, err := NewUserSearchCondition(&NewUserArgs{})
	if err == nil {
		t.Fatal("expected search condition error")
	}
}

func TestUserCanChangeNameAndEmail(
	t *testing.T,
) {
	user, err := ReconstructUser(&NewUserArgs{
		ID:    intPointer(1),
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err != nil {
		t.Fatalf("failed to reconstruct user: %v", err)
	}

	if err := user.Rename(stringPointer("bob")); err != nil {
		t.Fatalf("expected rename success, got: %v", err)
	}
	if err := user.ChangeEmail(stringPointer("bob@example.com")); err != nil {
		t.Fatalf("expected mail change success, got: %v", err)
	}

	if user.Name().GetValue() != "bob" {
		t.Fatalf("expected renamed user, got: %s", user.Name().GetValue())
	}
	if user.Email().GetValue() != "bob@example.com" {
		t.Fatalf("expected changed email, got: %s", user.Email().GetValue())
	}
}

func TestNewUserRejectsNameBlacklist(
	t *testing.T,
) {
	_, err := NewUser(&NewUserArgs{
		Name:          stringPointer("root user"),
		Email:         stringPointer("root@example.com"),
		NameBlacklist: []string{"root"},
	})
	if err == nil {
		t.Fatal("expected blacklist validation error")
	}
}

func TestUserRenameRejectsNameBlacklist(
	t *testing.T,
) {
	user, err := ReconstructUser(&NewUserArgs{
		ID:    intPointer(1),
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err != nil {
		t.Fatalf("failed to reconstruct user: %v", err)
	}

	if err := user.Rename(stringPointer("root user"), []string{"root"}); err == nil {
		t.Fatal("expected blacklist validation error")
	}
	if user.Name().GetValue() != "alice" {
		t.Fatalf("name should not change after failed rename, got: %s", user.Name().GetValue())
	}
}

func TestUserEnsureReadyToUpdateRequiresLifecycleState(
	t *testing.T,
) {
	user, err := NewUser(&NewUserArgs{
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := user.EnsureReadyToUpdate(); err == nil {
		t.Fatal("expected update lifecycle error")
	}
}

func TestUserEnsureReadyToCreateRequiresDraftWithNameAndEmail(
	t *testing.T,
) {
	validUser, err := NewUser(&NewUserArgs{
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err != nil {
		t.Fatalf("failed to create draft user: %v", err)
	}
	if err := validUser.EnsureReadyToCreate(); err != nil {
		t.Fatalf("expected create-ready user, got: %v", err)
	}

	persistedUser, err := ReconstructUser(&NewUserArgs{
		ID:    intPointer(1),
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err != nil {
		t.Fatalf("failed to reconstruct user: %v", err)
	}
	if err := persistedUser.EnsureReadyToCreate(); err == nil {
		t.Fatal("expected identity validation error")
	}

	missingName, err := NewUser(&NewUserArgs{Email: stringPointer("alice@example.com")})
	if err != nil {
		t.Fatalf("failed to create partial draft: %v", err)
	}
	if err := missingName.EnsureReadyToCreate(); err == nil {
		t.Fatal("expected name validation error")
	}
}

func TestReconstructUserListRequiresIdentityForEachUser(
	t *testing.T,
) {
	_, err := ReconstructUserList(&NewUserListArgs{
		Content: []NewUserArgs{
			{
				ID:    intPointer(1),
				Name:  stringPointer("alice"),
				Email: stringPointer("alice@example.com"),
			},
			{
				Name:  stringPointer("bob"),
				Email: stringPointer("bob@example.com"),
			},
		},
	})
	if err == nil {
		t.Fatal("expected identity error")
	}
}

func TestUserListCanManageCollectionRules(
	t *testing.T,
) {
	user, err := ReconstructUser(&NewUserArgs{
		ID:    intPointer(1),
		Name:  stringPointer("alice"),
		Email: stringPointer("alice@example.com"),
	})
	if err != nil {
		t.Fatalf("failed to reconstruct user: %v", err)
	}

	userList, err := ReconstructUserList(&NewUserListArgs{
		Content: []NewUserArgs{
			{
				ID:    intPointer(2),
				Name:  stringPointer("bob"),
				Email: stringPointer("bob@example.com"),
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to reconstruct user list: %v", err)
	}

	if userList.IsEmpty() {
		t.Fatal("expected user list not to be empty")
	}
	if userList.Count() != 1 {
		t.Fatalf("expected user count 1, got %d", userList.Count())
	}
	if err := userList.Append(*user); err != nil {
		t.Fatalf("expected append success, got: %v", err)
	}
	if !userList.ContainsIdentity(user.Identity()) {
		t.Fatal("expected appended identity")
	}
	if !userList.ContainsEmail(user.Email()) {
		t.Fatal("expected appended email")
	}
	if err := userList.Append(*user); err == nil {
		t.Fatal("expected duplicate user error")
	}
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
