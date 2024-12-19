package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/strcarne/spotify-downloader/internal/spotbackup"
	"github.com/strcarne/spotify-downloader/internal/spotmate"
	"github.com/strcarne/spotify-downloader/pkg/bytecalc"
)

var (
	ConvertThroughput  int
	DownloadThroughput int

	ConvertTimeout time.Duration
)

func main() {
	savePath := flag.String("save-path", ".", "path to save downloaded songs")
	convert := flag.Int("convert", 16, "how much convert tracks to mp3 at once")
	download := flag.Int("download", 16, "how much download tracks at once")
	showCreds := flag.Bool("show-creds", false, "show spotmate creds (dev purposes)")
	convertTimeoutStr := flag.String("convert-timeout", "16s", "how long to wait for song to convert (if <= 0, no timeout)")
	flag.Parse()

	var err error
	ConvertTimeout, err = time.ParseDuration(*convertTimeoutStr)
	if err != nil {
		fmt.Printf("Failed to parse given convert timeout [%s]: %s\n", *convertTimeoutStr, err)

		return
	}

	ConvertThroughput = *convert
	DownloadThroughput = *download

	args := flag.Args()

	creds, err := spotmate.GetCreds()
	if err != nil {
		log.Println(err)

		return
	}

	if *showCreds {
		fmt.Println(" --- Spotmate Creds ---")
		fmt.Println("CSRF:", creds.CSRF)
		fmt.Println("Cookies:", creds.Cookies)
		fmt.Println()
	}

	for _, arg := range args {
		switch {
		case strings.Contains(arg, "playlist"):
			fmt.Printf("Provided: %s\n", arg)
			fmt.Println(" --- Playlist --- ")

			handlePlaylist(arg, *savePath, creds)

		case strings.Contains(arg, "track"):
			fmt.Println(" --- Track --- ")
			fmt.Printf("Provided: %s\n", arg)

			handleTrack(arg, *savePath, creds)

		default:
			fmt.Println(" --- Unknown --- ")
			fmt.Printf("Provided: %s\n", arg)
			fmt.Println("Didn't recognize given item. Please provide playlist or track urls")
		}

		fmt.Println()
	}
}

func handlePlaylist(playlistURL string, savePath string, creds *spotmate.Creds) {
	infoURL, err := spotbackup.GetPlaylistInfoURL(playlistURL)
	if err != nil {
		fmt.Printf("Failed to fetch info url for playlist [%s]\nCause: %s\n", playlistURL, err)
		return
	}

	info, err := spotbackup.GetPlaylistInfo(infoURL)
	if err != nil {
		fmt.Printf("Failed to fetch info for playlist [%s]\nCause: %s\n", playlistURL, err)
		return
	}

	convertPipe := make(chan string, ConvertThroughput)
	downloadPipe := make(chan string, DownloadThroughput)

	go func() {
		defer close(convertPipe)

		for _, row := range info {
			fmt.Printf("Processing [%s] by [%s]\n", row[spotbackup.Song], row[spotbackup.Author])
			trackURL := row[spotbackup.TrackURL]
			convertPipe <- trackURL
		}
	}()

	go func() {
		defer close(downloadPipe)

		for trackURL := range convertPipe {
			resp, err := spotmate.Convert(trackURL, ConvertTimeout, creds)
			if err != nil {
				fmt.Printf("Failed to convert URL to MP3 [%s]\nCause: %s\n", trackURL, err)
				fmt.Println()

				continue
			}

			switch resp.Type() {
			case spotmate.RegularResponseType:
				r := resp.(spotmate.Response)
				if r.Error {
					fmt.Printf(
						"Failed to convert URL to MP3 [%s]\nCause: %s\n",
						trackURL,
						"something went wrong",
					)
					fmt.Println()

					continue
				}

				downloadPipe <- r.URL

			case spotmate.ErrorResponseType:
				e := resp.(spotmate.ErrorResponse)
				fmt.Printf("Failed to convert URL to MP3 [%s]\nCause: %s\n", trackURL, e.Error)
				fmt.Println()

			default:
				fmt.Printf(
					"Failed to convert URL to MP3 [%s]\nCause: %s\n",
					trackURL,
					"unknown response type",
				)
				fmt.Println()

			}
		}
	}()

	suspiciousDownloadedTracks := make([]*spotmate.DownloadedTrack, 0)

	for downloadURL := range downloadPipe {
		downloadedTrackInfo, err := spotmate.DownloadTrack(downloadURL, savePath)
		if err != nil {
			fmt.Printf("[download stage] ERROR: %s\n", err)
		}

		if downloadedTrackInfo.Size < 1.5*bytecalc.BytesInMB {
			suspiciousDownloadedTracks = append(suspiciousDownloadedTracks, downloadedTrackInfo)
		}
	}

	if len(suspiciousDownloadedTracks) > 0 {
		fmt.Println("WARNING!")
		fmt.Println("This downloaded tracks were marked as suspicious, because of size.")
		fmt.Println("To avoid any problems, please check them manually.")
		fmt.Println()

		for _, info := range suspiciousDownloadedTracks {
			fmt.Printf(
				"\tSong: %s\n\tSize: %s\n\tCallback URL: %s\n\tSave path: %s\n\n",
				info.Name,
				bytecalc.CalculateSizeLiteral(info.Size),
				info.CallbackURL,
				info.SavePath,
			)
		}
	}
}

func handleTrack(trackURL string, savePath string, creds *spotmate.Creds) {
	fmt.Println("Trying to download", trackURL)

	resp, err := spotmate.Convert(trackURL, ConvertTimeout, creds)
	if err != nil {
		fmt.Printf("Failed to convert URL to MP3 [%s]\nCause: %s\n", trackURL, err)
		fmt.Println()

		return
	}

	switch resp.Type() {
	case spotmate.RegularResponseType:
		r := resp.(spotmate.Response)
		if r.Error {
			fmt.Printf(
				"Failed to convert URL to MP3 [%s]\nCause: %s\n",
				trackURL,
				"something went wrong",
			)
			fmt.Println()

			return
		}

		downloadedTrackInfo, err := spotmate.DownloadTrack(r.URL, savePath)
		if err != nil {
			fmt.Printf("Failed to download [%s]: %s\n", trackURL, err)
		}

		if downloadedTrackInfo.Size < 1.5*bytecalc.BytesInMB {
			fmt.Println("WARNING!")
			fmt.Println("This downloaded tracks were marked as suspicious, because of size.")
			fmt.Println("To avoid any problems, please check them manually.")
			fmt.Println()

			fmt.Printf(
				"\tSong: %s\n\tSize: %s\n\tCallback URL: %s\n\tSave path: %s\n\n",
				downloadedTrackInfo.Name,
				bytecalc.CalculateSizeLiteral(downloadedTrackInfo.Size),
				downloadedTrackInfo.CallbackURL,
				downloadedTrackInfo.SavePath,
			)
		}

	case spotmate.ErrorResponseType:
		e := resp.(spotmate.ErrorResponse)
		fmt.Printf("Failed to convert URL to MP3 [%s]\nCause: %s\n", trackURL, e.Error)
		fmt.Println()

	default:
		fmt.Printf(
			"Failed to convert URL to MP3 [%s]\nCause: %s\n",
			trackURL,
			"unknown response type",
		)
		fmt.Println()

	}
}
