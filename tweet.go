package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"
)

type uploadMediaResponse struct {
	MediaID          int64     `json:"media_id"`
	MediaIDString    string    `json:"media_id_string"`
	Size             int       `json:"size"`
	ExpiresAfterSecs int       `json:"expires_after_secs"`
	Image            imageInfo `json:"image"`
}
type imageInfo struct {
	ImageType string `json:"image_type"`
	Width     int    `json:"w"`
	Height    int    `json:"h"`
}

func tweetWithImage(creds *creds, fileStr, message string) (*http.Response, error) {
	//boundaryBody作成
	var body bytes.Buffer
	mpWriter := multipart.NewWriter(&body)

	boundary := "END_OF_PART"
	if err := mpWriter.SetBoundary(boundary); err != nil {
		return nil, err
	}
	//part作成
	part := make(textproto.MIMEHeader)
	part.Set("Content-Disposition", "form-data; name=\"media_data\";")
	writer, err := mpWriter.CreatePart(part)
	if err != nil {
		return nil, err
	}
	//値(BASE64の画像バイナリを値にする)
	buffer, err := ioutil.ReadFile(fileStr)
	if err != nil {
		return nil, err
	}
	b64Buf := base64.StdEncoding.EncodeToString(buffer)
	writer.Write([]byte(b64Buf))

	mpWriter.Close()

	authHeader := manualOauthSettings(creds, map[string]string{}, "POST", UPLOADMEDIA)

	req, err := http.NewRequest("POST", UPLOADMEDIA, bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var res uploadMediaResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, err
	}

	resp, err = tweet(creds, message, []string{res.MediaIDString})
	if err != nil {
		return nil, err
	}

	return resp, nil

}

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
	contentType := http.DetectContentType(buffer)
	//読み取りポインタをリセットする
	file.Seek(0, 0)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	additionalParam := map[string]string{
		"command":     "INIT",
		"media_type":  contentType,
		"total_bytes": strconv.FormatInt(fileInfo.Size(), 10),
	}

	authHeader := manualOauthSettings(creds, additionalParam, "POST", UPLOADMEDIA)
	req, err := http.NewRequest("POST", UPLOADMEDIA, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)

	req.URL.RawQuery = sortedQueryString(additionalParam)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}

func tweet(creds *creds, message string, mediaIDs []string) (*http.Response, error) {
	addtionalParam := map[string]string{"status": message}
	if len(mediaIDs) != 0 {
		addtionalParam["media_ids"] = strings.Join(mediaIDs, ",")
	}
	authHeader := manualOauthSettings(creds, addtionalParam, "POST", UPDATESTATUS)

	req, err := http.NewRequest("POST", UPDATESTATUS, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.URL.RawQuery = sortedQueryString(addtionalParam)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp, nil
}
