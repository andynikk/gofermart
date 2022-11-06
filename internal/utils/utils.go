package utils

import (
	"bytes"
	"net/http"
	"strings"
)

func GETQuery(addressPost string) (*http.Response, error) {
	req, err := http.NewRequest("GET", addressPost, strings.NewReader(""))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func POSTQuery(addressPost string, message []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", addressPost, bytes.NewBuffer(message))
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
