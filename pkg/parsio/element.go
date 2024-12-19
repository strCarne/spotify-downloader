package parsio

import "golang.org/x/net/html"

func FindElement(startFrom *html.Node, tag string) *html.Node {
	if startFrom == nil {
		return nil
	}

	if startFrom.Data == tag {
		return startFrom
	}

	child := startFrom.FirstChild

	for startFrom.NextSibling != nil {
		startFrom = startFrom.NextSibling

		result := FindElement(startFrom, tag)
		if result != nil {
			return result
		}
	}

	return FindElement(child, tag)
}

func FindElementOfClass(startFrom *html.Node, tag, class string) *html.Node {
	elem := FindElement(startFrom, tag)
	if elem == nil {
		return nil
	}

	for _, attr := range elem.Attr {
		if attr.Key == "class" && attr.Val == class {
			return elem
		}
	}

	result := FindElementOfClass(elem.NextSibling, tag, class)
	if result != nil {
		return result
	}

	return FindElementOfClass(elem.FirstChild, tag, class)
}
