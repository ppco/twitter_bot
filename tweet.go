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

// tweetWithImage 画像つきツイート
func tweetWithImage(creds *creds, defaultImage, message string) (*http.Response, error) {
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
	//TODO: ここをAPI接続に変更する
	buffer, err := ioutil.ReadFile(defaultImage)
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

// tweetWithMedia 動画などのメディア付きツイート
// 1.Init
// 2.Append
// 3.Finalize
// の順に実行する
func tweetWithMedia(creds *creds, defaultMedia string) (*http.Response, error) {
	file, err := os.Open(defaultMedia)
	if err != nil {
		return nil, err
	}
	defer func() {
		file.Close()
	}()

	initRes, totalFileSize, err := mediaInit(creds, file)
	if err != nil {
		return nil, err
	}

	res, err := mediaAppend(creds, *initRes, totalFileSize, file)
	if err != nil {
		return nil, err
	}

	res, err = mediaFinalize(creds, *initRes)
	if err != nil {
		return nil, err
	}

	res, err = tweet(creds, "動画投稿テスト", []string{initRes.MediaIDString})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// mediaStatus メディア付きツイートのステータス取得
func mediaStatus(creds *creds, initRes uploadMediaResponse) (*http.Response, error) {
	param := map[string]string{
		"command":  "STATUS",
		"media_id": initRes.MediaIDString,
	}

	authHeader := manualOauthSettings(creds, param, "GET", UPLOADMEDIA)

	req, err := http.NewRequest("GET", UPLOADMEDIA, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.URL.RawQuery = sortedQueryString(param)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}

// mediaFinalize メディア付きツイートの終了処理
func mediaFinalize(creds *creds, initRes uploadMediaResponse) (*http.Response, error) {
	param := map[string]string{
		"command":  "FINALIZE",
		"media_id": initRes.MediaIDString,
	}

	authHeader := manualOauthSettings(creds, param, "POST", UPLOADMEDIA)

	req, err := http.NewRequest("POST", UPLOADMEDIA, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.URL.RawQuery = sortedQueryString(param)

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return resp, nil
}

// mediaAppend メディア付きツイートの追加処理
func mediaAppend(creds *creds, initRes uploadMediaResponse, totalFileSize int64, file *os.File) (*http.Response, error) {
	// TODO: チャンクにすると、「Segments do not add up to provided total file size.」が発生するため、現状一括でアップロードで対応
	chunked := make([]byte, totalFileSize)
	segmentIndex := 0
	var res *http.Response
	for {
		//boundaryBody作成
		var body bytes.Buffer
		mpWriter := multipart.NewWriter(&body)

		boundary := "END_OF_PART"
		if err := mpWriter.SetBoundary(boundary); err != nil {
			return nil, err
		}

		{
			//part作成(メディアデータ本体)
			part := make(textproto.MIMEHeader)
			part.Set("Content-Disposition", "form-data; name=\"media_data\";")
			writer, err := mpWriter.CreatePart(part)
			if err != nil {
				return nil, err
			}
			//指定バイト数だけチャンク
			n, err := file.Read(chunked)
			if n == 0 {
				break
			}
			if err != nil {
				return nil, err
			}

			b64Buf := base64.StdEncoding.EncodeToString(chunked)
			writer.Write([]byte(b64Buf))
		}

		{
			//その他パラメータの作成
			part := make(textproto.MIMEHeader)
			additionalParam := map[string]string{
				"command":       "APPEND",
				"media_id":      initRes.MediaIDString,
				"segment_index": strconv.Itoa(segmentIndex),
			}
			for k, v := range additionalParam {
				part.Set("Content-Disposition", fmt.Sprintf("form-data; name=\"%s\";", k))
				writer, err := mpWriter.CreatePart(part)
				if err != nil {
					return nil, err
				}
				writer.Write([]byte(v))
			}
		}

		mpWriter.Close()

		authHeader := manualOauthSettings(creds, map[string]string{}, "POST", UPLOADMEDIA)

		req, err := http.NewRequest("POST", UPLOADMEDIA, bytes.NewReader(body.Bytes()))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

		client := http.Client{}
		res, err := client.Do(req)
		if err != nil {
			return res, err
		}
		defer res.Body.Close()
		segmentIndex++
	}
	return res, nil
}

// mediaInit メディア付きツイートの初期化処理
func mediaInit(creds *creds, file *os.File) (*uploadMediaResponse, int64, error) {
	//fileからcontentTypeを読み取る
	buffer := make([]byte, 512)
	file.Read(buffer)
	contentType := http.DetectContentType(buffer)
	//読み取りポインタをリセットする
	file.Seek(0, 0)

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, 0, err
	}
	totalFileSize := fileInfo.Size()
	fmt.Println(totalFileSize)
	additionalParam := map[string]string{
		"command":     "INIT",
		"media_type":  contentType,
		"total_bytes": strconv.FormatInt(totalFileSize, 10),
	}

	authHeader := manualOauthSettings(creds, additionalParam, "POST", UPLOADMEDIA)
	req, err := http.NewRequest("POST", UPLOADMEDIA, nil)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", authHeader)

	req.URL.RawQuery = sortedQueryString(additionalParam)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	var res uploadMediaResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return nil, 0, err
	}
	return &res, totalFileSize, nil
}

// tweet ツイート処理
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
