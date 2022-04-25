package client

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
)

type SlackApp struct {
	Client http.Client
}

type errorResponse struct {
	Error string `json:"error"`
}

func New() SlackApp {
	return SlackApp{
		Client: http.Client{},
	}
}

func (a *SlackApp) Request(ctx context.Context, method string) ([]byte, error) {

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://slack.com/api/"+method, nil)
	if err != nil {
		return nil, err
	}
	result, err := a.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()
	resultBody, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}
	if result.StatusCode != http.StatusOK {
		return nil, errors.New(string(resultBody))
	}
	var errorJson errorResponse
	err = json.Unmarshal(resultBody, &errorJson)
	if err != nil {
		return nil, err
	}
	if errorJson.Error != "" {
		return nil, errors.New(errorJson.Error)
	}
	return resultBody, nil
}
