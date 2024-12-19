package parsio

import (
	"errors"
	"io"

	"github.com/strcarne/spotify-downloader/pkg/errwrap"
	"golang.org/x/net/html"
)

var ErrNoSuccessNoticeLink = errors.New("couldn't find success notice link")

func ParseSuccessNotice(page io.Reader) (string, error) {
	const location = "parsio.ParseSuccessNotice"

	root, err := html.Parse(page)
	if err != nil {
		return "", errwrap.Wrap(location, "couldn't parse html page", err)
	}

	result := findSuccessNotice(root)
	if result == nil {
		return "", errwrap.Wrap(location, "couldn't parse success notice", ErrNoSuccessNoticeLink)
	}

	return *result, nil
}

func findSuccessNotice(node *html.Node) *string {
	if node.Data == "div" {
		isSuccessNotice := false
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "success-notice" {
				isSuccessNotice = true
			}
		}

		if isSuccessNotice {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if c.Data == "a" {
					for _, attr := range c.Attr {
						if attr.Key == "href" {
							return &attr.Val
						}
					}
				}
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result := findSuccessNotice(child)
		if result != nil {
			return result
		}
	}

	return nil
}
