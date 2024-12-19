package spotmate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/strcarne/spotify-downloader/pkg/errwrap"
)

const (
	SpotmateURL = "https://spotmate.online"
	ConvertURL  = SpotmateURL + "/convert"
	UserAgent   = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
)

func Convert(trackURL string, timeout time.Duration, creds *Creds) (ConvertResponse, error) {
	const location = "spotmate.Convert"

	jsonData := bytes.NewBufferString(fmt.Sprintf(`{"urls": "%s"}`, trackURL))

	req, err := http.NewRequest("POST", ConvertURL, jsonData)
	if err != nil {
		return nil, errwrap.Wrap(location, "failed to create POST request", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", creds.CSRF)
	req.Header.Set("Origin", SpotmateURL)
	req.Header.Set("Referer", SpotmateURL+"/")
	req.Header.Set("User-Agent", UserAgent)

	for _, cookie := range creds.Cookies {
		req.AddCookie(cookie)
	}

	client := new(http.Client)

	if timeout.Milliseconds() > 0 {
		client.Timeout = timeout
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errwrap.Wrap(location, "http request failed", err)
	}
	defer resp.Body.Close()

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(resp.Body)

	response := Response{}
	if err := json.Unmarshal(buffer.Bytes(), &response); err != nil {
		errorResponse := ErrorResponse{}
		if err := json.Unmarshal(buffer.Bytes(), &errorResponse); err != nil {
			return nil, errwrap.Wrap(location, "failed to unmarshal", err)
		}
		return errorResponse, nil
	}

	return response, nil
}
