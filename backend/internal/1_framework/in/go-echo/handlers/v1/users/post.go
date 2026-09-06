package users

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"backend/internal/1_framework/in/go-echo/openapi"
	"backend/internal/2_adapter/controller"
	groupObject "backend/internal/4_domain/group_object"
	"backend/internal/logger"
)

func Post(
	echoContext echo.Context,
	toController controller.ToController,
) (
	err error,
) {
	request := openapi.UserCreate{}
	if err = echoContext.Bind(&request); err != nil {
		err = echoContext.JSON(
			http.StatusBadRequest,
			openapi.Error{
				Code:    http.StatusBadRequest,
				Message: "invalid request body",
			},
		)

		return
	}

	newUser, err := groupObject.NewUser(&groupObject.NewUserArgs{
		Name:  &request.Name,
		Email: &request.Email,
	})
	if err != nil {
		logger.Logging(echoContext.Request().Context(), err)
		err = echoContext.JSON(
			http.StatusBadRequest,
			openapi.Error{Code: http.StatusBadRequest, Message: err.Error()},
		)

		return
	}

	createdUser, err := toController.CreateUser(
		echoContext.Request().Context(),
		*newUser,
	)
	if err != nil {
		logger.Logging(echoContext.Request().Context(), err)
		err = echoContext.JSON(
			http.StatusBadRequest,
			openapi.Error{Code: http.StatusBadRequest, Message: err.Error()},
		)

		return
	}

	err = echoContext.JSON(
		http.StatusCreated,
		openapi.User{
			Id:    createdUser.ID().GetValue(),
			Name:  createdUser.Name().GetValue(),
			Email: createdUser.Email().GetValue(),
		},
	)

	return
}
