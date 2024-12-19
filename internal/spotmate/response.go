package spotmate

type ResponseType int

const (
	RegularResponseType ResponseType = iota
	ErrorResponseType
)

type ConvertResponse interface {
	Type() ResponseType
}

type Response struct {
	Error bool   `json:"error"`
	URL   string `json:"url"`
}

func (r Response) Type() ResponseType {
	return RegularResponseType
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (e ErrorResponse) Type() ResponseType {
	return ErrorResponseType
}
