package parsio

import (
	"errors"
	"io"

	"github.com/strcarne/spotify-downloader/pkg/errwrap"
	"golang.org/x/net/html"
)

var ErrNoMetaWithCSRF = errors.New("couldn't find meta with csrf")

func ParseCSRFFromMeta(page io.Reader) (string, error) {
	const location = "parsio.ParseCSRFFromMeta"

	root, err := html.Parse(page)
	if err != nil {
		return "", errwrap.Wrap(location, "couldn't parse html page", err)
	}

	token := findCSRF(root)

	if token == nil {
		return "", errwrap.Wrap(location, "couldn't parse csrf from meta", ErrNoMetaWithCSRF)
	}

	return *token, nil
}

func findCSRF(node *html.Node) *string {
	if node.Data == "meta" {
		isCSRF := false
		content := ""

		for _, attr := range node.Attr {
			if attr.Key == "content" {
				content = attr.Val
				if isCSRF {
					return &content
				}
			} else if attr.Key == "name" && attr.Val == "csrf-token" {
				isCSRF = true
				if content != "" {
					return &content
				}
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		token := findCSRF(child)
		if token != nil {
			return token
		}
	}

	return nil
}
