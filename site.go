package main

// Talking to the guild website:
//   POST /api/uploader/start  → a connect code to approve on the website, and a secret to poll with
//   POST /api/uploader/poll   → the key, once the member clicked Connect
//   POST /api/upload          → the saved code, with "Authorization: Bearer <key>"

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 30 * time.Second}

// errDisconnected means the key no longer works: the member disconnected it or left the guild.
var errDisconnected = errors.New("disconnected")

func post(site, path, key string, body any, out any) (int, error) {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", site+path, bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "FameAndFortuneUploader/"+version)
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if out != nil {
		_ = json.Unmarshal(data, out)
	}
	return res.StatusCode, nil
}

type startReply struct {
	Code      string `json:"code"`
	Poll      string `json:"poll"`
	URL       string `json:"url"`
	ExpiresIn int    `json:"expiresIn"`
}

func startConnect(site, name string) (startReply, error) {
	var r startReply
	status, err := post(site, "/api/uploader/start", "", map[string]string{"name": name}, &r)
	if err != nil {
		return r, err
	}
	if status != 200 || r.Poll == "" {
		return r, fmt.Errorf("the website said %d", status)
	}
	return r, nil
}

// waitForKey asks every 3 seconds until the code is approved or runs out.
func waitForKey(site string, s startReply, stop <-chan struct{}) (string, error) {
	deadline := time.Now().Add(time.Duration(s.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-stop:
			return "", errors.New("stopped")
		case <-time.After(3 * time.Second):
		}
		var r struct {
			Status string `json:"status"`
			Key    string `json:"key"`
		}
		if _, err := post(site, "/api/uploader/poll", "", map[string]string{"poll": s.Poll}, &r); err != nil {
			continue
		}
		switch r.Status {
		case "approved":
			return r.Key, nil
		case "expired":
			return "", errors.New("the connect code ran out")
		}
	}
	return "", errors.New("the connect code ran out")
}

type uploadReply struct {
	Error   string `json:"error"`
	Member  string `json:"member"`
	Results []struct {
		OK         bool   `json:"ok"`
		Status     string `json:"status"`
		Message    string `json:"message"`
		Characters []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"characters"`
		Todo []string `json:"todo"`
	} `json:"results"`
}

func upload(site, key string, codes []string, account string, at int64) (uploadReply, error) {
	var r uploadReply
	status, err := post(site, "/api/upload", key, map[string]any{"codes": codes, "account": account, "savedAt": at}, &r)
	if err != nil {
		return r, err
	}
	if status == 401 || status == 403 {
		return r, errDisconnected
	}
	if status != 200 {
		if r.Error == "" {
			r.Error = fmt.Sprintf("the website said %d", status)
		}
		return r, errors.New(r.Error)
	}
	return r, nil
}
