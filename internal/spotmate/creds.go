package spotmate

import (
	"net/http"

	"github.com/strcarne/spotify-downloader/pkg/errwrap"
	"github.com/strcarne/spotify-downloader/pkg/parsio"
)

type Creds struct {
	CSRF    string
	Cookies []*http.Cookie
}

func GetCreds() (*Creds, error) {
	const location = "spotmate.GetCreds"

	resp, err := http.Get("https://spotmate.online/")
	if err != nil {
		return nil, errwrap.Wrap(
			location,
			"unfortunately, spotmate.online is down:",
			err,
		)
	}
	defer resp.Body.Close()

	csrfToken, err := parsio.ParseCSRFFromMeta(resp.Body)
	if err != nil {
		return nil, errwrap.Wrap(location, "couldn't parse csrf token", err)
	}

	cookies := resp.Cookies()

	return &Creds{
		CSRF:    csrfToken,
		Cookies: cookies,
	}, nil
}
