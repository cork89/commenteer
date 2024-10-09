package snoo

import (
	"fmt"
	"log"
	c "main/common"
	"main/dataaccess"
	"strings"
)

type TemplateCache struct {
	Data map[c.RedditRequest]c.Link
}

func populateCache() {
	linkJson := dataaccess.GetLinks()
	for k, v := range linkJson {
		req := strings.Split(k, "-")
		bigReq := c.RedditRequest{Subreddit: req[0], Article: req[1], Comment: req[2]}
		linkCache[bigReq] = v
	}
	log.Println(linkCache)
}

func addToCache(req c.RedditRequest, link *c.Link) {
	dataaccess.AddLink(req, link, 0)
	linkCache[req] = *link
}

func UpdateCdnUrl(req c.RedditRequest, imageType string) {
	cdnUrl := fmt.Sprintf("%s/%s.%s", CdnBaseUrl, req.AsString(), imageType)
	dataaccess.UpdateCdnUrl(req, cdnUrl)
	link := linkCache[req]
	link.CdnUrl = cdnUrl
	linkCache[req] = link
}

func GetCache() map[c.RedditRequest]c.Link {
	// var values []c.Link
	// for _, value := range linkCache {
	// 	values = append(values, value)
	// }
	// return values
	// templateData := TemplateCache{Data: linkCache}
	// return templateData
	return linkCache
}
