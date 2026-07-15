package files

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/microsoft/UpdateAssistant/modules/config"
)

var client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

func base() string { return "https://" + config.Addr() + config.APIPrefix + "/attachments" }

func Upload(filename string, data []byte) (string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(data); err != nil {
		return "", err
	}
	w.Close()

	req, err := http.NewRequest(http.MethodPost, base(), &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+config.Key())
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("upload failed: " + resp.Status)
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func Download(id string) ([]byte, string, error) {
	req, err := http.NewRequest(http.MethodGet, base()+"/"+id, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+config.Key())

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", errors.New("download failed: " + resp.Status)
	}

	name := id
	cd := resp.Header.Get("Content-Disposition")
	if cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil {
			if fn, ok := params["filename"]; ok && fn != "" {
				name = fn
			}
		}
	}

	data, err := io.ReadAll(resp.Body)
	return data, name, err
}
