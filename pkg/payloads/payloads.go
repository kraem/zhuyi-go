package payloads

import (
	"fmt"

	"github.com/kraem/zhuyi-go/zettel"
)

type AppendRequest struct {
	Payload struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	} `json:"payload"`
}

type AppendResponse struct {
	Payload *appendPayload `json:"payload,omitempty"`
	Error   *string        `json:"error"`
}

type appendPayload struct {
	FileName *string `json:"file_name,omitempty"`
}

func NewAppendResponse(fn *string, err error) AppendResponse {
	var errStringPtr *string
	var payload *appendPayload
	if err != nil {
		es := fmt.Sprintf("%v", err.Error())
		errStringPtr = &es
	}
	payload = &appendPayload{
		FileName: fn,
	}
	return AppendResponse{
		Payload: payload,
		Error:   errStringPtr,
	}
}

type GraphResponse struct {
	Payload struct {
		Graph *zettel.D3jsGraph `json:"graph,omitempty"`
	} `json:"payload"`
	Error *string `json:"error"`
}

type IsolatedResponse struct {
	Payload struct {
		Zettels []zettel.Zettel `json:"unlinked_zettels"`
	} `json:"payload"`
	Error *string `json:"error"`
}
