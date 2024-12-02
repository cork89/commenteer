package dataaccess

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	c "main/common"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var dbpool *pgxpool.Pool

type Db struct{}

func createPool(postgresUrl string) {
	var err error
	if dbpool == nil {
		dbpool, err = pgxpool.New(context.Background(), postgresUrl)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func getConnection() (*pgx.Conn, error) {
	conn, err := dbpool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	return conn.Conn(), nil
}

func (d Db) GetRecentLinks(page int) (links []c.Link) {
	offset := (page - 1) * 10
	rows, err := dbpool.Query(context.Background(), `
SELECT l.cdn_image_url, l.user_id, l.image_url, l.query_id, l.link_id, l.cdn_image_height, l.cdn_image_width
FROM links l
WHERE l.cdn_image_url != '' ORDER BY l.created_date DESC LIMIT 10 OFFSET ($1);`, offset)
	// return handleRetrieve(rows, err)
	if err != nil {
		log.Printf("Error retrieving recent links, %v", err)
	}
	for rows.Next() {
		var link c.Link
		err := rows.Scan(&link.CdnUrl, &link.UserId, &link.ImageUrl, &link.QueryId, &link.LinkId, &link.ImageHeight, &link.ImageWidth)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return links
		}
		links = append(links, link)
	}
	return links
}

func (d Db) GetRecentLinksByUsername(page int, username string) (links []c.Link, userExists bool) {
	row := dbpool.QueryRow(context.Background(), `
SELECT u.user_id
FROM users u
WHERE u.username = ($1)`, username)
	var userId int
	err := row.Scan(&userId)
	if err != nil {
		log.Printf("User does not exist, %v", err)
		return nil, false
	}

	offset := (page - 1) * 10
	rows, err := dbpool.Query(context.Background(), `
SELECT l.cdn_image_url, l.user_id, l.image_url, l.query_id, l.link_id, l.cdn_image_height, l.cdn_image_width
FROM links l, users u
WHERE l.cdn_image_url != '' AND l.user_id = u.user_id AND u.user_id = ($1)
 ORDER BY l.created_date DESC LIMIT 10 OFFSET ($2);`, userId, offset)
	// return handleRetrieve(rows, err)
	if err != nil {
		log.Printf("Error retrieving recent links, %v", err)
		return nil, true
	}
	for rows.Next() {
		var link c.Link
		err := rows.Scan(&link.CdnUrl, &link.UserId, &link.ImageUrl, &link.QueryId, &link.LinkId, &link.ImageHeight, &link.ImageWidth)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return links, true
		}
		links = append(links, link)
	}
	return links, true
}

func (d Db) GetRecentLoggedInLinks(page int, userId int) (links []c.UserLinkData) {
	offset := (page - 1) * 10
	rows, err := dbpool.Query(context.Background(), `
select l.cdn_image_url, l.user_id, l.image_url, l.query_id, l.link_id, l.cdn_image_height, l.cdn_image_width, COALESCE(ua.active, false) AS active
from links l
LEFT JOIN useractions ua 
	ON ua.user_id = l.user_id AND ua.user_id = ($1) AND ua.target_id = l.link_id AND ua.action_type = 'like' AND ua.target_type = 'link'
WHERE l.cdn_image_url != ''
ORDER BY l.created_date DESC LIMIT 10 OFFSET ($2);`, userId, offset)

	if err != nil {
		log.Printf("Error retrieving recent links, %v", err)
	}
	for rows.Next() {
		var link c.Link
		var userAction c.UserAction
		err := rows.Scan(&link.CdnUrl, &link.UserId, &link.ImageUrl, &link.QueryId, &link.LinkId, &link.ImageHeight, &link.ImageWidth, &userAction.Active)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return links
		}
		links = append(links, c.UserLinkData{Link: link, UserAction: &userAction})
	}
	return links
}

func (d Db) GetRecentLoggedInSavedLinks(page int, userId int) (links []c.UserLinkData) {
	offset := (page - 1) * 10
	rows, err := dbpool.Query(context.Background(), `
select l.cdn_image_url, l.user_id, l.image_url, l.query_id, l.link_id, l.cdn_image_height, l.cdn_image_width, ua.active
from links l
INNER JOIN useractions ua 
	ON ua.user_id = ($1) AND ua.target_id = l.link_id AND ua.action_type = 'like' AND ua.target_type = 'link' AND ua.active = true
WHERE l.cdn_image_url != ''
ORDER BY l.created_date DESC LIMIT 10 OFFSET ($2);`, userId, offset)

	if err != nil {
		log.Printf("Error retrieving recent saved links, %v", err)
	}
	for rows.Next() {
		var link c.Link
		var userAction c.UserAction
		err := rows.Scan(&link.CdnUrl, &link.UserId, &link.ImageUrl, &link.QueryId, &link.LinkId, &link.ImageHeight, &link.ImageWidth, &userAction.Active)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return links
		}
		links = append(links, c.UserLinkData{Link: link, UserAction: &userAction})
	}
	return links
}

func (d Db) GetRecentLoggedInLinksByUsername(page int, userId int, username string) (links []c.UserLinkData) {
	offset := (page - 1) * 10
	rows, err := dbpool.Query(context.Background(), `
select l.cdn_image_url, l.user_id, l.image_url, l.query_id, l.link_id, l.cdn_image_height, l.cdn_image_width, COALESCE(ua.active, false) AS active
from users u, links l
LEFT JOIN useractions ua 
	ON ua.user_id = l.user_id AND ua.user_id = ($1) AND ua.target_id = l.link_id AND ua.action_type = 'like' AND ua.target_type = 'link'
WHERE l.cdn_image_url != '' AND l.user_id = u.user_id AND u.username = ($2)
ORDER BY l.created_date DESC LIMIT 10 OFFSET ($3);`, userId, username, offset)

	if err != nil {
		log.Printf("Error retrieving recent links, %v", err)
	}
	for rows.Next() {
		var link c.Link
		var userAction c.UserAction
		err := rows.Scan(&link.CdnUrl, &link.UserId, &link.ImageUrl, &link.QueryId, &link.LinkId, &link.ImageHeight, &link.ImageWidth, &userAction.Active)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return links
		}
		links = append(links, c.UserLinkData{Link: link, UserAction: &userAction})
	}
	return links
}

func (d Db) GetLinks() (links map[string]c.Link) {
	rows, err := dbpool.Query(context.Background(), `
SELECT l.image_url, l.proxy_url, l.query_id, l.cdn_image_url, c.comment, c.author, l.user_id, l.link_id, l.cdn_image_height, l.cdn_image_width
FROM links l, comments c
WHERE l.link_id = c.link_id
ORDER BY l.created_date
LIMIT 30`)
	return handleRetrieve(rows, err)
}

func (d Db) GetLink(req c.RedditRequest) (*c.Link, bool) {
	rows, err := dbpool.Query(context.Background(), `
SELECT l.image_url, l.proxy_url, l.query_id, l.cdn_image_url, c.comment, c.author, l.user_id, l.link_id, l.cdn_image_height, l.cdn_image_width
FROM links l, comments c
WHERE l.link_id = c.link_id and l.query_id = ($1)`, req.AsString())
	if err != nil {
		log.Printf("Query error: %v\n", err)
		return nil, false
	}
	links := handleRetrieve(rows, err)
	link, ok := links[req.AsString()]
	if !ok {
		return nil, false
	}

	return &link, true
}

func handleRetrieve(rows pgx.Rows, err error) (links map[string]c.Link) {
	links = make(map[string]c.Link)
	if err != nil {
		log.Printf("Query error: %v\n", err)
		return links
	}
	for rows.Next() {
		var queryId string
		var link c.Link
		var comment c.Comment
		err := rows.Scan(&link.ImageUrl, &link.ProxyUrl, &queryId, &link.CdnUrl, &comment.Comment, &comment.Author, &link.UserId, &link.LinkId, &link.ImageHeight, &link.ImageWidth)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return links
		}
		val, ok := links[queryId]
		if !ok {
			comments := []c.Comment{comment}
			link.RedditComments = comments
			link.QueryId = queryId
			links[queryId] = link
		} else {
			val.RedditComments = append(val.RedditComments, comment)
			links[queryId] = val
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("Row iteration error: %v\n", err)
	}
	return links
}

func (d Db) GetLoggedInLink(req c.RedditRequest, userId int) (*c.UserLinkData, bool) {
	rows, err := dbpool.Query(context.Background(), `
select l.cdn_image_url, l.user_id, l.image_url, l.query_id, l.link_id, l.cdn_image_height, l.cdn_image_width, COALESCE(ua.active, false) AS active
from links l
LEFT JOIN useractions ua 
	ON ua.user_id = l.user_id AND ua.target_id = l.link_id AND ua.action_type = 'like' AND ua.target_type = 'link'
WHERE l.user_id = ($1) and l.query_id = ($2) and l.cdn_image_url != '';`, userId, req.AsString())

	if err != nil {
		log.Printf("Error retrieving recent links, %v", err)
	}
	var userLinkData *c.UserLinkData
	for rows.Next() {
		var link c.Link
		var userAction c.UserAction
		err := rows.Scan(&link.CdnUrl, &link.UserId, &link.ImageUrl, &link.QueryId, &link.LinkId, &link.ImageHeight, &link.ImageWidth, &userAction.Active)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return userLinkData, false
		}
		userLinkData = &c.UserLinkData{Link: link, UserAction: &userAction}
	}
	return userLinkData, true
}

func (d Db) AddLink(req c.RedditRequest, link *c.Link, userId int) {
	var linkId int
	query := "INSERT INTO links (image_url, proxy_url, query_id, cdn_image_url, user_id) VALUES ($1, $2, $3, '', $4) RETURNING link_id"
	args := []any{link.ImageUrl, link.ProxyUrl, req.AsString(), userId}
	err := dbpool.QueryRow(context.Background(), query, args[0], args[1], args[2], args[3]).Scan(&linkId)

	if err != nil {
		log.Printf("error inserting link: %v\n", err)
		return
	}

	_, err = dbpool.CopyFrom(
		context.Background(),
		pgx.Identifier{"comments"},
		[]string{"link_id", "comment", "author"},
		pgx.CopyFromSlice(len(link.RedditComments), func(i int) ([]any, error) {
			return []any{linkId, link.RedditComments[i].Comment, link.RedditComments[i].Author}, nil
		}),
	)
	if err != nil {
		log.Printf("error inserting comments: %v\n", err)
	}
}

func (d Db) UpdateCdnUrl(req c.RedditRequest, cdnUrl string, height int, width int) {
	query := "UPDATE links SET cdn_image_url = ($1), cdn_image_height = ($2), cdn_image_width = ($3) WHERE query_id = ($4)" // AND cdn_image_url IS DISTINCT FROM ($5)"
	args := []any{cdnUrl, req.AsString(), height, width}
	_, err := dbpool.Exec(context.Background(), query, args[0], args[2], args[3], args[1])

	if err != nil {
		log.Printf("error updating link: %v, cdnUrl: %s, err: %v\n", req, cdnUrl, err)
		return
	}
}

func (d Db) GetUser(username string) (*c.User, bool) {
	// var userCookie c.UserCookie
	var user c.User
	query := "SELECT username,subscribed,refresh_token,refresh_expire_dt_tm,icon_url,access_token,user_id,remaining_uploads,upload_refresh_dt_tm from users where username = ($1)"
	err := dbpool.QueryRow(context.Background(), query, username).Scan(&user.Username, &user.Subscribed, &user.RefreshToken, &user.RefreshExpireDtTm, &user.IconUrl, &user.AccessToken, &user.UserId, &user.RemainingUploads, &user.UploadRefreshDtTm)

	if err != nil {
		log.Println("no user found, ", err)
		return nil, false
	}
	// user.UserCookie = userCookie
	return &user, true
}

func (d Db) AddUser(user c.User) bool {
	slog.Info("AddUser", "username", user.Username, "refresh_token", user.RefreshToken, "refresh_expire_dt_tm", user.RefreshExpireDtTm, "icon_url", user.IconUrl, "access_token", user.AccessToken)
	query := "INSERT INTO users (username, refresh_token, refresh_expire_dt_tm, icon_url, access_token) VALUES ($1, $2, $3, $4, $5)"
	args := []any{user.Username, user.RefreshToken, user.RefreshExpireDtTm, user.IconUrl, user.AccessToken}
	_, err := dbpool.Exec(context.Background(), query, args[0], args[1], args[2], args[3], args[4])

	if err != nil {
		log.Printf("error inserting user: %v\n", err)
		return false
	}
	return true
}

func (d Db) UpdateUser(username string, accessToken string, refreshExpireDtTm time.Time) bool {
	query := "UPDATE users SET access_token = $1, refresh_expire_dt_tm = $2 where username = ($3)"
	args := []any{accessToken, refreshExpireDtTm, username}
	_, err := dbpool.Exec(context.Background(), query, args[0], args[1], args[2])
	if err != nil {
		log.Printf("error updating user: %v\n", err)
		return false
	}
	return true
}

func (d Db) DecrementUserUploadCount(userId int) bool {
	query := "UPDATE users SET remaining_uploads = remaining_uploads - 1 where user_id = ($1)"
	args := []any{userId}
	_, err := dbpool.Exec(context.Background(), query, args[0])
	if err != nil {
		log.Printf("error updating user: %v\n", err)
		return false
	}
	return true
}

func (d Db) RefreshUserUploadCount(userId int, newCount int) bool {
	query := "UPDATE users SET remaining_uploads = ($1), upload_refresh_dt_tm = NOW() + INTERVAL '1 week' where user_id = ($2) and remaining_uploads < ($3)"
	args := []any{userId, newCount}
	_, err := dbpool.Exec(context.Background(), query, args[1], args[0], args[1])
	if err != nil {
		log.Printf("error updating user: %v\n", err)
		return false
	}
	return true
}

func (d Db) AddUserAction(userAction c.UserAction) bool {
	query := `
INSERT INTO useractions (user_id, action_type, target_id, target_type) VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id, action_type, target_id, target_type)
DO UPDATE SET active = NOT useractions.active`
	args := []any{userAction.UserId, userAction.ActionType, userAction.TargetId, userAction.TargetType}
	_, err := dbpool.Exec(context.Background(), query, args[0], args[1], args[2], args[3])

	if err != nil {
		log.Printf("error inserting useraction: %v\n", err)
		return false
	}
	return true
}

func (d Db) GetLinkId(queryId string) (linkId int, err error) {
	row := dbpool.QueryRow(context.Background(), `
SELECT l.link_id
FROM links l
WHERE l.query_id = ($1)`, queryId)
	err = row.Scan(&linkId)
	return linkId, err
}

func (d Db) AddLinkStyles(linkStyles []c.LinkStyle) bool {
	linkId, err := d.GetLinkId(linkStyles[0].QueryId)

	if err != nil {
		log.Printf("failed to retrieve link id, %v", err)
		return false
	}

	query := `
INSERT INTO linkstyle (link_id, style_key, style_value)
VALUES %s
ON CONFLICT (link_id, style_key) DO UPDATE
SET style_value = EXCLUDED.style_value`

	values := []string{}
	args := []interface{}{}
	argIdx := 1

	for _, linkStyle := range linkStyles {
		values = append(values, fmt.Sprintf("($%d, $%d, $%d)", argIdx, argIdx+1, argIdx+2))
		args = append(args, linkId, linkStyle.Key, linkStyle.Value)
		argIdx += 3
	}

	fullQuery := fmt.Sprintf(query, strings.Join(values, ","))

	_, err = dbpool.Exec(context.Background(), fullQuery, args...)

	if err != nil {
		log.Printf("Failed to add link styles, err=%v\n", err)
		return false
	}
	return true
}

func (d Db) GetLinkStyles(linkId int) (linkStyles []c.LinkStyle, err error) {
	rows, err := dbpool.Query(context.Background(), `
SELECT ls.link_style_id, ls.link_id, ls.style_key, ls.style_value
FROM linkstyle ls
WHERE ls.link_id = ($1);`, linkId)

	if err != nil {
		log.Printf("Error retrieving linkstyles for link_id=%d, err= %v", linkId, err)
	}
	linkStyles = make([]c.LinkStyle, 0)
	for rows.Next() {
		var linkStyle c.LinkStyle
		err := rows.Scan(&linkStyle.LinkStyleId, &linkStyle.LinkId, &linkStyle.Key, &linkStyle.Value)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return linkStyles, err
		}
		linkStyles = append(linkStyles, linkStyle)
	}
	return linkStyles, nil
}

func (d Db) InitializeDb() {
	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPw := os.Getenv("POSTGRES_PASSWORD")
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresDb := os.Getenv("POSTGRES_DB")
	postgresOpt := os.Getenv("POSTGRES_OPT")

	postgresUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s%s", postgresUser, postgresPw, postgresHost, postgresPort, postgresDb, postgresOpt)
	createPool(postgresUrl)
}
