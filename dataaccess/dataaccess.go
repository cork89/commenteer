package dataaccess

import (
	c "main/common"
	"time"
)

var dataAccess DataAccess

type DataAccess interface {
	GetLinks() map[string]c.Link
	GetRecentLinks(page int) []c.Link
	GetRecentLoggedInLinks(page int, userId int) []c.UserLinkData
	GetLink(req c.RedditRequest) (*c.Link, bool)
	GetLoggedInLink(req c.RedditRequest, userId int) (*c.UserLinkData, bool)
	AddLink(req c.RedditRequest, link *c.Link, userId int)
	UpdateCdnUrl(req c.RedditRequest, cdnUrl string, height int, width int)
	GetUser(username string) (*c.User, bool)
	AddUser(user c.User) bool
	UpdateUser(username string, accessToken string, refreshExpireDtTm time.Time) bool
	DecrementUserUploadCount(userId int) bool
	RefreshUserUploadCount(userId int, newCount int) bool
	AddUserAction(userAction c.UserAction) bool
}

func GetLinks() (linkJson map[string]c.Link) {
	return dataAccess.GetLinks()
}

func GetRecentLinks(page int) []c.Link {
	return dataAccess.GetRecentLinks(page)
}

func GetRecentLoggedInLinks(page int, userId int) (links []c.UserLinkData) {
	return dataAccess.GetRecentLoggedInLinks(page, userId)
}

func GetLink(req c.RedditRequest) (*c.Link, bool) {
	return dataAccess.GetLink(req)
}

func GetLoggedInLink(req c.RedditRequest, userId int) (*c.UserLinkData, bool) {
	return dataAccess.GetLoggedInLink(req, userId)
}

func AddLink(req c.RedditRequest, link *c.Link, userId int) {
	dataAccess.AddLink(req, link, userId)
}

func UpdateCdnUrl(req c.RedditRequest, cdnUrl string, height int, width int) {
	dataAccess.UpdateCdnUrl(req, cdnUrl, height, width)
}

func GetUser(username string) (*c.User, bool) {
	return dataAccess.GetUser(username)
}

func AddUser(user c.User) bool {
	return dataAccess.AddUser(user)
}

func UpdateUser(username string, accessToken string, refreshExpireDtTm time.Time) bool {
	return dataAccess.UpdateUser(username, accessToken, refreshExpireDtTm)
}

func DecrementUserUploadCount(userId int) bool {
	return dataAccess.DecrementUserUploadCount(userId)
}

func RefreshUserUploadCount(userId int, newCount int) bool {
	return dataAccess.RefreshUserUploadCount(userId, newCount)
}

func AddUserAction(userAction c.UserAction) bool {
	return dataAccess.AddUserAction(userAction)
}

func Initialize(dataAccessType string) {
	if dataAccessType == "local" {
		dataAccess = &Local{}
	} else {
		dataAccess = &Db{}

		migrator, err := NewMigrator()
		if err != nil {
			panic(err)
		}

		now, exp, info, err := migrator.Info()
		if err != nil {
			panic(err)
		}

		if now < exp {
			// migration is required, dump out the current state
			// and perform the migration
			println("migration needed, current state:")
			println(info)

			err = migrator.Migrate()
			if err != nil {
				panic(err)
			}
			println("migration successful!")
		} else {
			println("no database migration needed")
		}
	}
}
