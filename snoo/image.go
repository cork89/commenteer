package snoo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
)

const (
	maxWidth  int = 1024
	halfWidth int = 512
)

type ImgOpts struct {
	dstX int
	dstY int
	padX int
}

func resizeImage() ImgOpts {
	return ImgOpts{dstX: maxWidth, dstY: 0, padX: 0}
}

func resizeImageTODO(img string) ImgOpts {
	// log.Printf("img: %s", img)
	res, err := http.Get(img)
	if err != nil {
		log.Printf("failed retrieving image, %v\n", err)
		return ImgOpts{dstX: maxWidth, dstY: 0, padX: 0}
	}
	defer res.Body.Close()

	// buff := new(bytes.Buffer)

	// _, err = io.Copy(buff, res.Body)

	// if err != nil {
	// 	log.Printf("failed to read to byte buffer: %v\n", err)
	// }

	// src, err := jpeg.Decode(res.Body)
	src, _, err := image.Decode(res.Body)

	if err != nil {
		log.Printf("failed to decode image: %v\n", err)
		return ImgOpts{dstX: maxWidth, dstY: 0, padX: 0}
	}
	srcX := src.Bounds().Max.X
	srcY := src.Bounds().Max.Y

	dstX := srcX
	dstY := srcY

	if srcX > maxWidth || srcY > maxWidth {
		if srcY > srcX*2 {
			dstY = srcY * (halfWidth / srcX)
			dstX = halfWidth
		} else {
			dstY = srcY * (maxWidth / srcX)
			dstX = maxWidth
		}
	} else {
		if srcX > srcY {
			dstX = maxWidth
			dstY = maxWidth - (srcY/srcX)*maxWidth
		} else {
			dstY = maxWidth
			dstX = maxWidth - (srcX/srcY)*maxWidth
		}
	}
	padX := maxWidth - dstX

	return ImgOpts{dstX: dstX, dstY: dstY, padX: padX}
}

func computeHMACSHA256(url string, key string) ([]byte, error) {
	h := hmac.New(sha256.New, []byte(key))

	_, err := h.Write([]byte(url))
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func GetImgProxyUrl(src string) (string, error) {
	imgOpts := resizeImage()
	imgproxy := os.Getenv("IMGPROXY_URL")
	imgProxySalt := os.Getenv("IMGPROXY_SALT")
	imgProxyKey := os.Getenv("IMGPROXY_KEY")

	// res, err := http.Get("http://" + imgproxy + ":8080/preset:sharp/resize:fit:700/plain/https://i.redd.it/b7zui0ibi3p91.jpg@jpg")
	args := fmt.Sprintf("/resize:fit:%d:0:1/padding:0:%d/wm:1:soea:0:0:0.5/background:255:255:255/plain/%s", imgOpts.dstX, imgOpts.padX, src)

	rslt, err := computeHMACSHA256(fmt.Sprintf("%s%s", imgProxySalt, args), imgProxyKey)

	if err != nil {
		return "", err
	}

	base64 := base64.RawURLEncoding.EncodeToString(rslt)

	return fmt.Sprintf("http://%s/%s%s", imgproxy, base64, args), nil
}
