package deepseek_api

import (
	"encoding/json"
	"net/http"
)

func (th *Client) ListModels() (reply ListModelsReply, err error) {
	req, err := th.NewRequest(http.MethodGet, ModelRoute, nil)
	if err != nil {
		return
	}

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = findReplyError(res.StatusCode)
		return
	}

	err = json.NewDecoder(res.Body).Decode(&reply)
	return
}

type ListModelsReply struct {
	Object string      `json:"object"`
	Data   []ModelData `json:"data"`
}

type ModelData struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}
