package main

import (
	"fmt"
	"net/http"
)

func main() {
	cred := getCreds()

	//resp, err := tweetWithMedia(cred, "gopher_ueda.png")
	resp, err := tweet(cred, "てsつとテウと")
	if err != nil {
		fmt.Println(err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp)
	}
}
