package GoToREST

import (
	"encoding/json"
	"github.com/rs/cors"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	GET      = "GET"  //X
	POST     = "POST" //X
	PUT      = "PUT"  //X
	PATCH    = "PATCH"
	DELETE   = "DELETE" //x
	COPY     = "COPY"
	HEAD     = "HEAD"
	OPTIONS  = "OPTIONS"
	LINK     = "LINK"
	UNLINK   = "UNLINK"
	PURGE    = "PURGE"
	LOCK     = "LOCK"
	UNLOCK   = "UNLOCK"
	PROPFIND = "PROPFIND"
	VIEW     = "VIEW"
)

type RestService func([]byte /*requestData*/, map[string]string /*pathParam*/, map[string]string /*queryParam*/, map[string]string /*postParam*/, map[string]string /*headerParam*/, any) (any, RestServiceStatus)

type RestServiceInterceptor func([]byte /*requestData*/, map[string]string /*pathParam*/, map[string]string /*queryParam*/, map[string]string /*postParam*/, map[string]string /*headerParam*/) (*InterceptorError, any)

type RestServiceStatus struct {
	StatusCode int
	Message    string
}

type InterceptorError struct {
	StatusCode int
	Message    string
}

func (error *InterceptorError) Error() string {
	println("INTERCEPTOR error: " + error.Message)
	return error.Message
}

func (d RestServiceStatus) Error() string {
	println("REST service error: " + d.Message)
	return d.Message
}

type RestServer struct {
	serverMutex *http.ServeMux
}

type Mapping struct {
	mapping        map[string] /*path*/ map[string] /*method*/ RestService
	mappingPathVar map[string] /*path*/ map[string] /*method*/ []string /*path vars key value*/
}

type restError struct{ error }

var globalMapping = Mapping{}
var pathVarMatcher = regexp.MustCompile("\\{(.*?)\\}")
var pathEndMatcher = regexp.MustCompile("/$")
var pathVarExtractor = regexp.MustCompile("\\{(.*?)\\}/")

func NewRestServer() RestServer {
	return RestServer{serverMutex: http.NewServeMux()}
}

func (a RestServer) ListenAndServe() {
	listenAddr := ":9081"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}

	a.serverMutex.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		internalPath := r.URL.Path
		if isPathVarMapping(internalPath, r.Method) {
			internalPath = extractPathFromPathVarMapping(r.URL.Path, r.Method)
		}

		if globalMapping.mapping[internalPath][r.Method] == nil {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		created, remoteError := a.processRequest(r, globalMapping.mapping[internalPath][r.Method])

		if remoteError.StatusCode != 200 {
			w.WriteHeader(remoteError.StatusCode)
			w.Write([]byte(remoteError.Error()))
			return
		}

		if remoteError = a.writeResponse(created, w); remoteError.StatusCode != 200 {
			w.WriteHeader(remoteError.StatusCode)
			w.Write([]byte(remoteError.Error()))
		}
	})

	handler := cors.AllowAll().Handler(a.serverMutex)
	_ = http.ListenAndServe(listenAddr, handler)
}

func (a RestServer) HandleServiceIntercepted(method string, path string, restFunc RestService, interceptor RestServiceInterceptor) {
	a.HandleService(method, path, func(requestData []byte, pathParam map[string]string, queryParam map[string]string, postParam map[string]string, headerParam map[string]string, optional any) (any, RestServiceStatus) {

		err, session := interceptor(requestData, pathParam, queryParam, postParam, headerParam)
		if err != nil {
			return nil, RestServiceStatus{StatusCode: err.StatusCode, Message: err.Error()}
		}
		if session == nil {
			return nil, RestServiceStatus{StatusCode: http.StatusForbidden, Message: "Forbidden by request interceptor"}
		}

		return restFunc(requestData, pathParam, queryParam, postParam, headerParam, session)
	})
}

func (a RestServer) HandleService(method string, path string, restFunc RestService) {
	if globalMapping.mapping == nil {
		globalMapping.mapping = map[string]map[string]RestService{}

	}
	if globalMapping.mappingPathVar == nil {
		globalMapping.mappingPathVar = map[string]map[string][]string{}
	}

	var pathVars []string
	if pathVarMatcher.MatchString(path) { //check if path vars are part of url
		if !pathEndMatcher.MatchString(path) { //url have to end with "/"
			path = path + "/"
		}

		//extract path vars and remove surrounding brackets and slash
		pathVars = pathVarExtractor.FindAllString(path, -1)
		for pathVar := range pathVars {
			pathVars[pathVar] = strings.ReplaceAll(pathVars[pathVar], "{", "")
			pathVars[pathVar] = strings.ReplaceAll(pathVars[pathVar], "}/", "")
		}

		//extract url to bind (so remove all path vars)
		path = pathVarExtractor.ReplaceAllString(path, "([A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9]))/")
	}

	var newPath = false
	if globalMapping.mapping[path] == nil {
		newPath = true
		globalMapping.mapping[path] = map[string]RestService{}
	}
	if globalMapping.mappingPathVar[path] == nil {
		globalMapping.mappingPathVar[path] = map[string][]string{}
	}

	globalMapping.mapping[path][method] = restFunc
	globalMapping.mappingPathVar[path][method] = pathVars

	if newPath {
		a.serverMutex.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			internalPath := r.URL.Path
			if isPathVarMapping(internalPath, r.Method) {
				internalPath = extractPathFromPathVarMapping(r.URL.Path, r.Method)
			}

			if globalMapping.mapping[internalPath][r.Method] == nil {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			created, remoteError := a.processRequest(r, globalMapping.mapping[internalPath][r.Method])

			if remoteError.StatusCode != 200 {
				w.WriteHeader(remoteError.StatusCode)
				w.Write([]byte(remoteError.Error()))
				return
			}

			if remoteError = a.writeResponse(created, w); remoteError.StatusCode != 200 {
				w.WriteHeader(remoteError.StatusCode)
				w.Write([]byte(remoteError.Error()))
			}
		})
	}
}

func (a RestServer) processRequest(r *http.Request, restFunc RestService) (any, RestServiceStatus) {
	//extract query parameters
	err := r.ParseForm()
	reqParam := map[string]string{}
	for key, value := range r.Form {
		reqParam[key] = value[0]
	}

	//extract form-data parameters
	err = r.ParseMultipartForm(262144)
	postParam := map[string]string{}
	for key, value := range r.PostForm {
		postParam[key] = value[0]
	}

	//extract path variable values from url
	pathVarsMap, err := extractVarsFromPathVarMapping(r.URL.Path, r.Method)
	if err != nil {
		return nil, RestServiceStatus{StatusCode: http.StatusNotFound, Message: "could not be found"}
	}

	//extract header parameters
	headerParam := map[string]string{}
	for name, values := range r.Header { // Loop over all header
		for _, value := range values { // Loop over all values for the name.
			headerParam[name] = value
		}
	}

	//request object
	var responseAny any

	//read and unmarshal request
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, RestServiceStatus{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	//if no post parameters expect and handle json
	if len(postParam) <= 0 && len(body) > 0 {
		err = json.Unmarshal(body, &responseAny)
		if err != nil {
			return nil, RestServiceStatus{StatusCode: http.StatusBadRequest, Message: err.Error()}
		}

		jsonResponse, err := json.Marshal(responseAny)
		if err != nil {
			return nil, RestServiceStatus{StatusCode: http.StatusBadRequest, Message: err.Error()}
		}
		return restFunc(jsonResponse, pathVarsMap, reqParam, postParam, headerParam, nil) //call remote function
	} else {
		return restFunc(body, pathVarsMap, reqParam, postParam, headerParam, nil) //call remote function
	}
}

func isPathVarMapping(requestPath string, method string) bool {
	if !pathEndMatcher.MatchString(requestPath) { //url have to end with "/"
		requestPath = requestPath + "/"
	}

	for key, element := range globalMapping.mappingPathVar {
		matches := regexp.MustCompile(key)
		if matches.MatchString(requestPath) && element[method] != nil {
			return true
		}
	}
	return false
}

func extractPathFromPathVarMapping(requestPath string, method string) string {
	if !pathEndMatcher.MatchString(requestPath) { //url have to end with "/"
		requestPath = requestPath + "/"
	}

	for key, element := range globalMapping.mappingPathVar {
		matches := regexp.MustCompile(key)
		if matches.MatchString(requestPath) && element[method] != nil {
			return key
		}
	}
	return ""
}

func extractVarsFromPathVarMapping(requestPath string, method string) (map[string]string, error) {
	if !pathEndMatcher.MatchString(requestPath) { //url have to end with "/"
		requestPath = requestPath + "/"
	}

	pathVars := map[string]string{}
	mappedPath := extractPathFromPathVarMapping(requestPath, method)
	for key, element := range globalMapping.mappingPathVar {
		matches := regexp.MustCompile(key)
		if matches.MatchString(requestPath) && element[method] != nil {
			varsPosition := strings.Split(mappedPath, "/")
			urlParts := strings.Split(requestPath, "/")

			var varsValue []string
			for pathKey, pathVar := range varsPosition {
				if pathVar == "([A-Za-z0-9]([A-Za-z0-9-]*[A-Za-z0-9]))" && urlParts[pathKey] != "" {
					varsValue = append(varsValue, urlParts[pathKey])
				}
			}

			if len(globalMapping.mappingPathVar[key][method]) != len(varsValue) {
				return nil, restError{}
			}

			for pathVar := range globalMapping.mappingPathVar[key][method] {
				pathVars[globalMapping.mappingPathVar[key][method][pathVar]] = varsValue[pathVar]
			}
		}
	}
	return pathVars, nil
}

func (a RestServer) writeResponse(response any, w http.ResponseWriter) RestServiceStatus {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return RestServiceStatus{StatusCode: http.StatusInternalServerError, Message: err.Error()}
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

	return RestServiceStatus{StatusCode: http.StatusOK}
}
