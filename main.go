package main

import (
	"flag"
	"fmt"
	"net/http"
)

func main() {
	cred := getCreds()

	tweetType := flag.String("tweettype", "2", "tweettype")
	message := flag.String("message", "猫\n#猫", "defaultMessage")
	image := flag.String("image", "default_cats.jpeg", "defaultImage")
	media := flag.String("media", "default_cats.mp4", "defaultMedia")
	flag.Parse()

	var resp *http.Response
	var err error
	switch *tweetType {
	case TweetTypeText:
		resp, err = tweet(cred, *message, nil)
	case TweetTypeImg:
		resp, err = tweetWithImage(cred, *image, *message)
	case TweetTypeMedia:
		resp, err = tweetWithMedia(cred, *message, *media)
	}
	if err != nil {
		fmt.Printf("[ERROR] ツイッターボットでエラー発生：%v \n", err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[WARN] ツイッターボットでHTTPステータスが200ではない HTTPSTATUS:%d", resp.StatusCode)
	}
}
