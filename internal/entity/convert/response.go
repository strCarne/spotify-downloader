package convert

type Response struct {
	Error bool   `json:"error"`
	URL   string `json:"url"`
}
