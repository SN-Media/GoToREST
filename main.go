package main

import (
	"GoToREST/GoToREST"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	restServer := GoToREST.NewRestServer()

	println("Create info handler")
	restServer.HandleService(GoToREST.GET, "/api/createinfo", CreateInfoGETcall())
	restServer.HandleService(GoToREST.GET, "/{test}/api/{test2}/createinfoPathVar/{username}/{password}", CreateInfoGETcallPathVar()) //{username}/{password} //{test}
	restServer.HandleService(GoToREST.GET, "/api/createinfoQueryParamData", CreateInfoGETQueryparamdata())
	restServer.HandleService(GoToREST.POST, "/api/createinfoPOSTjson", CreateInfoPOSTjson())
	restServer.HandleService(GoToREST.POST, "/api/createinfoPOSTformdata", CreateInfoPOSTformdata())

	restServer.HandleServiceIntercepted(GoToREST.GET, "/api/intercepted", CreateInfoGETcall(), InterceptorTest())

	println("Start listening for REST calls")
	restServer.ListenAndServe()
}

// InterceptorTest
// Interceptor example
// *
func InterceptorTest() GoToREST.RestServiceInterceptor {
	return func(requestData []byte, pathParam map[string]string, queryParam map[string]string, postParam map[string]string, headerParam map[string]string) (*GoToREST.InterceptorError, any) {
		fmt.Println("InterceptorTest")
		return nil, false
	}
}

// CreateInfoGETcall
// GET call example
// *
func CreateInfoGETcall() GoToREST.RestService {
	return func(requestData []byte, pathParam map[string]string, queryParam map[string]string, postParam map[string]string, headerParam map[string]string, optional any) (any, GoToREST.RestServiceStatus) {
		fmt.Println("CreateInfoGETcall service was called")
		response := InfoResponse{}
		response.Userid = "u123456"
		response.Anrede = "Herr"
		response.Nachname = "Friedensfrau2"
		response.Vorname = "Sascha2"
		return response, GoToREST.RestServiceStatus{StatusCode: http.StatusOK}
	}
}

// CreateInfoGETcallPathVar
// GET call with path variables
// *
func CreateInfoGETcallPathVar() GoToREST.RestService {
	return func(requestData []byte, pathParam map[string]string, queryParam map[string]string, postParam map[string]string, headerParam map[string]string, optional any) (any, GoToREST.RestServiceStatus) {
		fmt.Println("CreateInfoGETcallPathVar service was called")

		fmt.Println("Path request param")
		for key, value := range pathParam {
			fmt.Println(key + ": " + value)
		}

		response := [2]InfoResponse{}
		response[0].Userid = "u123456"
		response[0].Anrede = "Herr"
		response[0].Nachname = "Friedensfrau1"
		response[0].Vorname = "Sascha1"
		response[1].Userid = "u654321"
		response[1].Anrede = "Frau"
		response[1].Nachname = "Friedensfrau2"
		response[1].Vorname = "Sascha2"
		return response, GoToREST.RestServiceStatus{StatusCode: http.StatusOK}
	}
}

// CreateInfoGETQueryparamdata
// GET call with query parameters
// *
func CreateInfoGETQueryparamdata() GoToREST.RestService {
	return func(requestData []byte, pathParam map[string]string, queryParam map[string]string, postParam map[string]string, headerParam map[string]string, optional any) (any, GoToREST.RestServiceStatus) {
		fmt.Println("CreateInfoGETcall service was called")

		fmt.Println("Query request param")
		for key, value := range queryParam {
			fmt.Println(key + ": " + value)
		}

		response := [2]InfoResponse{}
		response[0].Userid = "u123456"
		response[0].Anrede = "Herr"
		response[0].Nachname = "Friedensfrau1"
		response[0].Vorname = "Sascha1"
		response[1].Userid = "u654321"
		response[1].Anrede = "Frau"
		response[1].Nachname = "Friedensfrau2"
		response[1].Vorname = "Sascha2"
		return response, GoToREST.RestServiceStatus{StatusCode: http.StatusOK}
	}
}

// CreateInfoPOSTjson
// POST call giving json
// *
func CreateInfoPOSTjson() GoToREST.RestService {
	return func(requestData []byte, pathParam map[string]string, queryParam map[string]string, postParam map[string]string, headerParam map[string]string, optional any) (any, GoToREST.RestServiceStatus) {
		fmt.Println("CreateInfoPOSTjson service was called")

		request := InfoRequest{}
		err := json.Unmarshal(requestData, &request)
		if err != nil { //ignore err if unknown elements should be ignored
			return nil, GoToREST.RestServiceStatus{StatusCode: http.StatusBadRequest, Message: "Unmarshalling of request json failed"}
		}

		response := InfoResponse{}

		//TODO: implement me
		response.Userid = "u123456"
		response.Anrede = "Herr"
		response.Nachname = "Friedensfrau"
		response.Vorname = "Sascha"

		if 5 != 5 {
			remoteError := GoToREST.RestServiceStatus{StatusCode: http.StatusBadRequest, Message: "Some thing went wrong"}
			return nil, remoteError
		}

		return response, GoToREST.RestServiceStatus{StatusCode: http.StatusOK}
	}
}

// CreateInfoPOSTformdata
// POST call with form data
// *
func CreateInfoPOSTformdata() GoToREST.RestService {
	return func(requestData []byte, pathParam map[string]string, queryParam map[string]string, postParam map[string]string, headerParam map[string]string, optional any) (any, GoToREST.RestServiceStatus) {
		fmt.Println("CreateInfoPOSTjson service was called")
		response := InfoResponse{}

		fmt.Println("Post request param")
		for key, value := range postParam {
			fmt.Println(key + ": " + value)
		}

		//TODO: implement me
		response.Userid = "u123456"
		response.Anrede = "Herr"
		response.Nachname = "Friedensfrau"
		response.Vorname = "Sascha"

		if 5 != 5 {
			remoteError := GoToREST.RestServiceStatus{StatusCode: http.StatusBadRequest, Message: "Some thing went wrong"}
			return nil, remoteError
		}

		return response, GoToREST.RestServiceStatus{StatusCode: http.StatusOK}
	}
}

// InfoRequest
// Request struct for examples
type InfoRequest struct {
	Userid     string    `json:"userid"`
	Anrede     string    `json:"anrede"`
	Title      string    `json:"title"`
	Vorname    string    `json:"vorname"`
	Nachname   string    `json:"nachname"`
	Exported   bool      `json:"exported"`
	KartenNr   string    `json:"kartennr"`
	VertragsNr string    `json:"vertragsnr"`
	Tarif      string    `json:"tarif"`
	Mobilnr    string    `json:"mobilnr"`
	Email      string    `json:"email"`
	Plz        string    `json:"plz"`
	Ort        string    `json:"ort"`
	StrHausNr  string    `json:"strhausnr"`
	Timestamp  time.Time `json:"ts"`
}

// InfoResponse
// Response struct for examples
type InfoResponse struct {
	Userid     string    `json:"userid"`
	Anrede     string    `json:"anrede"`
	Title      string    `json:"title"`
	Vorname    string    `json:"vorname"`
	Nachname   string    `json:"nachname"`
	Exported   bool      `json:"exported"`
	KartenNr   string    `json:"kartennr"`
	VertragsNr string    `json:"vertragsnr"`
	Tarif      string    `json:"tarif"`
	Mobilnr    string    `json:"mobilnr"`
	Email      string    `json:"email"`
	Plz        string    `json:"plz"`
	Ort        string    `json:"ort"`
	StrHausNr  string    `json:"strhausnr"`
	Timestamp  time.Time `json:"ts"`
}
