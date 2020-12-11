package main

import (
	"fmt"
	"net/http"
)

func main() {
	cred := getCreds()

	//resp, err := tweetWithImage(cred, "gopher_ueda.png", "画像投稿")
	resp, err := tweetWithMedia(cred, "mov_hts-samp001.mp4")
	//resp, err := tweet(cred, "順番が重要だったのか")
	if err != nil {
		fmt.Println(err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp)
	}
}
