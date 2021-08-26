package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// ++
// +
// Dans ce fichier, on implemente quelques fonctions permettant d'envoyer des requetes HTTP a un service externe.
// Nous avons par exemple la fonction PostRequestToAPI qui permet d'envoyer une requete POST a un API.
// +
// ++

func GetRequestToAPI(requestUrl string) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic "+BASIC_AUTH_CREDENTIALS)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func PostRequestToAPI(requestUrl string, credentials string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, _ := http.NewRequest("POST", requestUrl, bytes.NewReader(data))
	req.Header.Add("Content-Type", API_CONTENT_TYPE)
	req.Header.Add("Authorization", "Basic "+credentials)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func PutRequestToAPI(requestUrl string, body interface{}) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, _ := http.NewRequest("PUT", requestUrl, bytes.NewReader(data))
	req.Header.Add("Content-Type", API_CONTENT_TYPE)
	req.Header.Add("Authorization", "Basic "+BASIC_AUTH_CREDENTIALS)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func DeleteRequestToAPI(requestUrl string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", requestUrl, nil)
	req.Header.Add("Authorization", "Basic "+BASIC_AUTH_CREDENTIALS)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
