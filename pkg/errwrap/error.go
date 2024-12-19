package errwrap

import (
	"encoding/json"
)

type Error struct {
	Location string `json:"location"`
	Message  string `json:"message"`
	Cause    error  `json:"cause"`
}

func (e Error) Error() string {

	serialized, err := json.MarshalIndent(&e, "", "  ")
	if err != nil {
		panic(err)
	}

	return string(serialized)
}

func Wrap(location string, message string, cause error) Error {
	return Error{Location: location, Message: message, Cause: cause}
}
