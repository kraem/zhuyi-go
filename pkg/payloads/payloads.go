package payloads

import (
	"fmt"

	"github.com/kraem/zhuyi-go/network"
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
		Graph *network.D3jsGraph `json:"graph,omitempty"`
	} `json:"payload"`
	Error *string `json:"error"`
}

type UnlinkedResponse struct {
	Payload struct {
		Nodes []network.Node `json:"unlinked_nodes"`
	} `json:"payload"`
	Error *string `json:"error"`
}

type DelRequest struct {
	Payload struct {
		FileName string `json:"file_name"`
	} `json:"payload"`
}

type DelResponse struct {
	Error *string `json:"error"`
}

type StatusResponse struct {
	Payload struct {
		Status string `json:"status"`
	} `json:"payload"`
}
