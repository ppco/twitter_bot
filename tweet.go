package main

import (
	"net/http"
	"net/url"
	"os"
	"strconv"
)

func tweetWithMedia(creds *creds, fileStr string) (*http.Response, error) {
	file, err := os.Open(fileStr)
	if err != nil {
		return nil, err
	}
	defer func() {
		file.Close()
	}()

	//fileからcontentTypeを読み取る
	buffer := make([]byte, 512)
	file.Read(buffer)
	//contentType := http.DetectContentType(buffer)
	//読み取りポインタをリセットする
	file.Seek(0, 0)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	authHeader := manualOauthSettings(creds, map[string]string{}, "", "")
	req, err := http.NewRequest("POST", UPLOADMEDIA, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	query := url.Values{}
	query.Add("command", "INIT")
	query.Add("total_bytes", strconv.FormatInt(fileInfo.Size(), 10))
	query.Add("media_type", "image/png")
	req.URL.RawQuery = "command=INIT&total_bytes=6962&media_type=image/png" //query.Encode()

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func tweet(creds *creds, message string) (*http.Response, error) {
	authHeader := manualOauthSettings(creds, map[string]string{"status": message}, "POST", UPDATESTATUS)

	req, err := http.NewRequest("POST", UPDATESTATUS, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	query := url.Values{}
	query.Add("status", message)
	req.URL.RawQuery = query.Encode()

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}
