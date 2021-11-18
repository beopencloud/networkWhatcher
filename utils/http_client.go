package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
)

// ++
// +
// Dans ce fichier, on implemente quelques fonctions permettant d'envoyer des requetes HTTP a un service externe.
// Nous avons par exemple la fonction PostRequestToAPI qui permet d'envoyer une requete POST a un API.
// +
// ++

func PostRequestToAPI(body interface{}) (*http.Response, error) {
	log.Println("POST Request body: ", body)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	credentials := base64.StdEncoding.EncodeToString([]byte(USERNAME + ":" + PASSWORD))
	client := &http.Client{}
	req, _ := http.NewRequest("POST", SERVICE_CREATE_EVENT_URL, bytes.NewReader(data))
	req.Header.Add("Content-Type", API_CONTENT_TYPE)
	req.Header.Add("Authorization", "Basic "+credentials)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func PutRequestToAPI(body interface{}) (*http.Response, error) {
	log.Println("PUT Request body: ", body)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	credentials := base64.StdEncoding.EncodeToString([]byte(USERNAME + ":" + PASSWORD))
	req, _ := http.NewRequest("PUT", SERVICE_UPDATE_EVENT_URL, bytes.NewReader(data))
	req.Header.Add("Content-Type", API_CONTENT_TYPE)
	req.Header.Add("Authorization", "Basic "+credentials)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func DeleteRequestToAPI(body interface{}) (*http.Response, error) {
	log.Println("DELETE Request body: ", body)
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	req, _ := http.NewRequest("DELETE", SERVICE_DELETE_EVENT_URL, bytes.NewReader(data))
	req.Header.Add("Authorization", "Basic "+BASIC_AUTH_CREDENTIALS)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
