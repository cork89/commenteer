package dataaccess

import (
	"io"
	"log"
	c "main/common"
	"time"
)

var dataAccess DataAccess
var bucketUploader BucketUploader

type DataAccess interface {
	GetLinks() map[string]c.Link
	GetRecentLinks(page int) []c.Link
	GetRecentLinksByUsername(page int, username string) ([]c.Link, bool)
	GetRecentLoggedInLinks(page int, userId int) []c.UserLinkData
	GetRecentLoggedInSavedLinks(page int, userId int) (links []c.UserLinkData)
	GetRecentLoggedInLinksByUsername(page int, userId int, username string) (links []c.UserLinkData)
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
	AddLinkStyles(linkStyles []c.LinkStyle) bool
	GetLinkStyles(linkId int) (linkStyles []c.LinkStyle, err error)
	InitializeDb()
}

func GetLinks() (linkJson map[string]c.Link) {
	return dataAccess.GetLinks()
}

func GetRecentLinks(page int) []c.Link {
	return dataAccess.GetRecentLinks(page)
}

func GetRecentLinksByUsername(page int, username string) ([]c.Link, bool) {
	return dataAccess.GetRecentLinksByUsername(page, username)
}

func GetRecentLoggedInLinks(page int, userId int) (links []c.UserLinkData) {
	return dataAccess.GetRecentLoggedInLinks(page, userId)
}

func GetRecentLoggedInSavedLinks(page int, userId int) (links []c.UserLinkData) {
	return dataAccess.GetRecentLoggedInSavedLinks(page, userId)
}

func GetRecentLoggedInLinksByUsername(page int, userId int, username string) (links []c.UserLinkData) {
	return dataAccess.GetRecentLoggedInLinksByUsername(page, userId, username)
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

func AddLinkStyles(linkStyles []c.LinkStyle) bool {
	return dataAccess.AddLinkStyles(linkStyles)
}

func GetLinkStyles(linkId int) (linkStyles []c.LinkStyle, err error) {
	return dataAccess.GetLinkStyles(linkId)
}

type BucketUploader interface {
	UploadImage(img io.Reader, fileName string) (*string, error)
	InitializeBucket()
}

func UploadImage(img io.Reader, fileName string) (*string, error) {
	return bucketUploader.UploadImage(img, fileName)
}

func Initialize(dataAccessType string) {
	if dataAccessType == "local" {
		dataAccess = &Local{}
		dataAccess.InitializeDb()
		bucketUploader = &LocalBucketUploader{}
		bucketUploader.InitializeBucket()
	} else {
		dataAccess = &Db{}
		dataAccess.InitializeDb()
		bucketUploader = &RealBucketUploader{}
		bucketUploader.InitializeBucket()
		go func() {
			migrator, err := NewMigrator()
			if err != nil {
				log.Printf("Failed to create migrator, err=%v\n", err)
				return
			}

			now, exp, info, err := migrator.Info()
			if err != nil {
				log.Printf("Failed to get migrator info, err=%v\n", err)
				return
			}

			if now < exp {
				// migration is required, dump out the current state
				// and perform the migration
				println("migration needed, current state:")
				println(info)

				err = migrator.Migrate()
				if err != nil {
					log.Printf("Failed to get migrate, err=%v\n", err)
					return
				}
				println("migration successful!")
			} else {
				println("no database migration needed")
			}
		}()
	}
}
