package spotmate

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/strcarne/spotify-downloader/pkg/bytecalc"
	"github.com/strcarne/spotify-downloader/pkg/errwrap"
)

const TryCount = 5

var (
	ErrUnexpectedStatusCode = errors.New("unexpected status code [expect 200]")
	ErrCouldntFetchSongName = errors.New("couldn't fetch song name")
)

// Returns size in bytes and error if any
func DownloadTrack(callbackURL string, savePath string) (*DownloadedTrack, error) {
	const location = "spotmate.DownloadTrack"

	resp, err := http.Get(callbackURL)
	for i := 0; err != nil && strings.Contains(err.Error(), "reset") && i < TryCount; i++ {
		time.Sleep(time.Millisecond * 100)
		resp, err = http.Get(callbackURL)
	}

	if err != nil {
		return nil, errwrap.Wrap(location, "failed to do GET request", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errwrap.Wrap(location, "request to callback URL failed", ErrUnexpectedStatusCode)
	}

	disposition := resp.Header.Get("Content-Disposition")
	songName, err := getSongName(disposition)
	if err != nil {
		return nil, errwrap.Wrap(location, "no song name in Content-Disposition", err)
	}

	if err := os.MkdirAll(savePath, 0755); err != nil && !os.IsExist(err) {
		return nil, errwrap.Wrap(location, fmt.Sprintf("failed to make save path [%s]", savePath), err)
	}

	songSavePath := filepath.Join(savePath, songName)
	if _, err := os.Stat(songSavePath); err != nil && !os.IsNotExist(err) {
		return nil, errwrap.Wrap(location, fmt.Sprintf("song [%s] already exists\n\n", songName), nil)
	}

	buffer := new(bytes.Buffer)
	buffer.ReadFrom(resp.Body)

	if err := os.WriteFile(songSavePath, buffer.Bytes(), 0644); err != nil {
		return nil, errwrap.Wrap(location, fmt.Sprintf("failed to save [%s] song into file", songName), err)
	}

	fmt.Printf(
		"Downloaded\n\tName: %s\n\tTo: %s\n\tTotal size: %s\n\n",
		songName,
		savePath,
		bytecalc.CalculateSizeLiteral(buffer.Len()),
	)

	return &DownloadedTrack{
		Name:        strings.TrimSuffix(songName, filepath.Ext(songName)),
		SavePath:    songSavePath,
		CallbackURL: callbackURL,
		Size:        buffer.Len(),
	}, nil
}

func getSongName(disposition string) (string, error) {
	const location = "spotmate.getSongName"

	values := strings.Split(disposition, ";")

	for _, value := range values {
		value = strings.TrimSpace(value)

		result := strings.Split(value, "=")
		if len(result) != 2 {
			continue
		}

		if strings.TrimSpace(strings.ToLower(result[0])) == "filename" {
			return strings.Trim(strings.TrimSpace(result[1]), `"`), nil
		}
	}

	return "", errwrap.Wrap(location, "disposition don't contain song name", ErrCouldntFetchSongName)
}
