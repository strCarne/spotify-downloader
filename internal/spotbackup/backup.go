package spotbackup

import (
	"encoding/csv"
	"net/http"
	"net/url"

	"github.com/strcarne/spotify-downloader/pkg/errwrap"
	"github.com/strcarne/spotify-downloader/pkg/parsio"
)

const BackupAPIProvider = "https://www.spotify-backup.com"

func GetPlaylistInfo(infoURL string) ([][]string, error) {
	const location = "spotbackup.GetPlaylistInfo"

	resp, err := http.Get(infoURL)
	if err != nil {
		return nil, errwrap.Wrap(location, "failed to get playlist info", err)
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	if reader == nil {
		return nil, errwrap.Wrap(location, "failed to parse playlist info", err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, errwrap.Wrap(location, "failed to parse playlist info", err)
	}

	return records, nil
}

func GetPlaylistInfoURL(playlistURL string) (string, error) {
	const location = "spotbackup.GetPlaylistInfoURL"

	values := url.Values{}
	values.Add("playlist_id", playlistURL)
	values.Add("terms", "on")
	values.Add("submit", "Generate Backup")

	resp, err := http.PostForm(BackupAPIProvider, values)
	if err != nil {
		return "", errwrap.Wrap(location, "failed to post form", err)
	}
	defer resp.Body.Close()

	path, err := parsio.ParseSuccessNotice(resp.Body)
	if err != nil {
		return "", errwrap.Wrap(location, "failed to parse success notice from body", err)
	}

	return BackupAPIProvider + path, nil
}
