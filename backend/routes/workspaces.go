package routes

import (
	"log"
	"pocketvue/constants"
	"pocketvue/helpers"

	"github.com/pocketbase/pocketbase/core"
)

func GetAllWorkspaces(e *core.RequestEvent) error {
	// Get authenticated user
	user, err := helpers.GetAuthenticatedUser(e)
	if err != nil {
		return err
	}

	// Log user information
	log.Printf("GetAllWorkspaces called by user: ID=%s, Email=%s, Name=%s",
		user.GetString("id"),
		user.GetString("email"),
		user.GetString("name"))

	records, err := helpers.FindAllRecords(e.App, constants.CollectionWorkspaces)
	if err != nil {
		return helpers.JSONInternalServerError(e, err.Error())
	}

	return helpers.JSONSuccess(e, records)
}
