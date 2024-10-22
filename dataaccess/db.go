package dataaccess

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	c "main/common"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

var dbpool *pgxpool.Pool

type Db struct{}

func getConnection(postgresUrl string) {
	var err error
	if dbpool == nil {
		dbpool, err = pgxpool.New(context.Background(), postgresUrl)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (d Db) GetRecentLinks(page int) (links map[string]c.Link) {
	offset := (page - 1) * 10
	rows, err := dbpool.Query(context.Background(), `
SELECT l.image_url, l.proxy_url, l.query_id, l.cdn_image_url, c.comment, c.author, l.user_id
FROM links l, comments c
WHERE l.link_id in (SELECT l2.link_id FROM links l2 ORDER BY l2.created_date DESC LIMIT 10 OFFSET ($1)) 
and l.link_id = c.link_id;`, offset)
	return handleRetrieve(rows, err)
}

func (d Db) GetLinks() (links map[string]c.Link) {
	rows, err := dbpool.Query(context.Background(), `
SELECT l.image_url, l.proxy_url, l.query_id, l.cdn_image_url, c.comment, c.author, l.user_id
FROM links l, comments c
WHERE l.link_id = c.link_id
ORDER BY l.created_date
LIMIT 30`)
	return handleRetrieve(rows, err)
}

func (d Db) GetLink(req c.RedditRequest) (*c.Link, bool) {
	rows, err := dbpool.Query(context.Background(), `
SELECT l.image_url, l.proxy_url, l.query_id, l.cdn_image_url, c.comment, c.author, l.user_id
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
		err := rows.Scan(&link.ImageUrl, &link.ProxyUrl, &queryId, &link.CdnUrl, &comment.Comment, &comment.Author, &link.UserId)
		if err != nil {
			log.Printf("Scan error: %v\n", err)
			return links
		}
		val, ok := links[queryId]
		if !ok {
			comments := []c.Comment{comment}
			link.RedditComments = comments
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

func (d Db) UpdateCdnUrl(req c.RedditRequest, cdnUrl string) {
	query := "UPDATE links SET cdn_image_url = ($1) WHERE query_id = ($2)"
	args := []any{cdnUrl, req.AsString()}
	err := dbpool.QueryRow(context.Background(), query, args[0], args[1])

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
	query := "UPDATE users SET remaining_uploads = ($1), upload_refresh_dt_tm = NOW() + INTERVAL '1 week' where user_id = ($2)"
	args := []any{newCount, userId}
	_, err := dbpool.Exec(context.Background(), query, args[0], args[1])
	if err != nil {
		log.Printf("error updating user: %v\n", err)
		return false
	}
	return true
}

func init() {
	err := godotenv.Load("/run/secrets/.env.local")
	if err != nil {
		log.Println(err)
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	postgresPw := os.Getenv("POSTGRES_PASSWORD")
	postgresHost := os.Getenv("POSTGRES_HOST")
	postgresPort := os.Getenv("POSTGRES_PORT")
	postgresDb := os.Getenv("POSTGRES_DB")
	postgresOpt := os.Getenv("POSTGRES_OPT")

	postgresUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s%s", postgresUser, postgresPw, postgresHost, postgresPort, postgresDb, postgresOpt)
	getConnection(postgresUrl)
}
