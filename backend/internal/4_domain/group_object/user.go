package group_object

import (
	"fmt"

	typeObject "backend/internal/4_domain/type_object"
)

type User struct {
	id    typeObject.ID
	name  typeObject.Name
	email typeObject.Email
}

type NewUserArgs struct {
	ID            *int
	Name          *string
	Email         *string
	NameBlacklist []string
}

func NewUser(
	args *NewUserArgs,
) (
	user *User,
	err error,
) {
	user = &User{}
	if args == nil {
		args = &NewUserArgs{}
	}

	user.id, err = typeObject.NewID(args.ID)
	if err != nil {
		user = nil
		return
	}

	user.name, err = typeObject.NewName(args.Name, args.NameBlacklist)
	if err != nil {
		user = nil
		return
	}

	user.email, err = typeObject.NewEmail(args.Email)
	if err != nil {
		user = nil
		return
	}

	return
}

func NewUserSearchCondition(
	args *NewUserArgs,
) (
	user *User,
	err error,
) {
	user, err = NewUser(args)
	if err != nil {
		user = nil

		return
	}
	if !user.CanBeUsedAsSearchCondition() {
		user, err = nil, fmt.Errorf("user search condition is required")

		return
	}

	err = nil

	return
}

func ReconstructUser(
	args *NewUserArgs,
) (
	user *User,
	err error,
) {
	user, err = NewUser(args)
	if err != nil {
		user = nil

		return
	}
	if !user.HasIdentity() {
		user, err = nil, fmt.Errorf("user identity is required")

		return
	}

	err = nil

	return
}

func (receiver User) ID() (
	iD typeObject.ID,
) {
	iD = receiver.id

	return
}

func (receiver User) Identity() (
	iD typeObject.ID,
) {
	iD = receiver.id

	return
}

func (receiver User) Name() (
	name typeObject.Name,
) {
	name = receiver.name

	return
}

func (receiver User) Email() (
	email typeObject.Email,
) {
	email = receiver.email

	return
}

func (receiver User) HasIdentity() (
	hasIdentity bool,
) {
	hasIdentity = receiver.id.GetValue() > 0

	return
}

func (receiver User) HasName() (
	hasName bool,
) {
	hasName = !receiver.name.GetIsNil() && receiver.name.GetValue() != ""

	return
}

func (receiver User) HasEmail() (
	hasEmail bool,
) {
	hasEmail = !receiver.email.GetIsNil() && receiver.email.GetValue() != ""

	return
}

func (receiver User) CanBeUsedAsSearchCondition() (
	canBeUsedAsSearchCondition bool,
) {
	canBeUsedAsSearchCondition = receiver.HasName() || receiver.HasEmail()

	return
}

func (receiver User) EnsureReadyToUpdate() (
	err error,
) {
	if !receiver.HasIdentity() {
		err = fmt.Errorf("user identity is required")

		return
	}
	if !receiver.HasName() {
		err = fmt.Errorf("user name is required")

		return
	}
	if !receiver.HasEmail() {
		err = fmt.Errorf("user email is required")

		return
	}

	err = nil

	return
}

func (receiver User) EnsureReadyToCreate() (
	err error,
) {
	if receiver.HasIdentity() {
		err = fmt.Errorf("new user must not have identity")

		return
	}
	if !receiver.HasName() {
		err = fmt.Errorf("user name is required")

		return
	}
	if !receiver.HasEmail() {
		err = fmt.Errorf("user email is required")

		return
	}

	err = nil

	return
}

func (receiver *User) Rename(
	value *string,
	nameBlacklist ...[]string,
) (
	err error,
) {
	blacklist := []string{}
	if len(nameBlacklist) > 0 {
		blacklist = nameBlacklist[0]
	}

	name, err := typeObject.NewName(value, blacklist)
	if err != nil {

		return
	}
	receiver.name = name

	err = nil

	return
}

func (receiver User) ValidateNameBlacklist(
	nameBlacklist []string,
) (
	err error,
) {
	value := receiver.name.GetValue()
	if receiver.name.GetIsNil() {
		err = nil

		return
	}

	_, err = typeObject.NewName(&value, nameBlacklist)

	return
}

func (receiver *User) ChangeEmail(
	value *string,
) (
	err error,
) {
	email, err := typeObject.NewEmail(value)
	if err != nil {

		return
	}
	receiver.email = email

	err = nil

	return
}
