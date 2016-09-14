package session

import (
	"fmt"
	"net/http"
	"sync"
	"time"
	"errors"
	"strings"

	"github.com/gorilla/sessions"
	"golang.org/x/net/context"

	"io/ioutil"
	"encoding/json"
	"bytes"
	"github.com/shampur/etcdstore"
)

//Service Interface of session manager
type Service interface {
	login(ctx context.Context, req LoginRequest) (LoginResponse, error)
	logout(ctx context.Context, req LogoutRequest) (LogoutResponse, error)
	validateapp(ctx context.Context, req validateAppRequest) (LoginResponse, error)
	apiprocess(ctx context.Context, req apiRequest)  (interface{}, error)
}

//validate app request
type validateAppRequest struct {
	httpreq *http.Request
}


type apiRequest struct {
	httpreq *http.Request
	data interface{}
}

//LogoutRequest
type LogoutRequest struct {
	httpreq *http.Request
}

// LoginRequest service
type LoginRequest struct {
	httpreq *http.Request
	cred    Credentials
}

// LogoutResponse service
type LogoutResponse struct {
	Session       *sessions.Session `json:"session"`
	Httpreq       *http.Request     `json:"httpreq"`
}
// LoginResponse service
type LoginResponse struct {
	Authenticated bool              `json:"authenticated"`
	Message       string            `json:"message"`
	Username      string		`json:"username"`
	Session       *sessions.Session `json:"session"`
	Httpreq       *http.Request     `json:"httpreq"`
}

//Credentials object of the user
type Credentials struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Organization string `json:"organization"`
}

type sessionService struct {
	mtx         	sync.RWMutex
	store       	*etcdstore.EtcdStore
	authmanager 	*AuthManager
	apiconfig	*apiConfig
}

type apiresponse struct {
	sessresponse	LoginResponse
	result 		interface{}
}


//NewSessionService contains the session store
func NewSessionService() Service {
	return &sessionService{
		store:       	etcdstore.NewEtcdStore([]string{"http://127.0.0.1:2379"}, "contivSession", []byte("something-very-secret")),
		authmanager: 	NewAuthmanager(),
		apiconfig: 	GetApiConfig(),
	}
}

func (s *sessionService) login(ctx context.Context, r LoginRequest) (LoginResponse, error) {
	fmt.Println("Login service called")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var res LoginResponse
	session, err := s.store.Get(r.httpreq, "contiv-session")

	if err != nil {
		fmt.Println("error while retrieving session info")
		return LoginResponse{}, err
	}

	if !session.IsNew {
		res, err = validate(session)
		if err != nil {
			return LoginResponse{}, err
		}
		if res.Authenticated != true {
			res, err = s.authmanager.authenticate(r.cred)
		}
	} else {
		res, err = s.authmanager.authenticate(r.cred)
	}
	if res.Authenticated {
		session.Values["Username"] = r.cred.Username
		session.Values["LastLoginTime"] = time.Now().Format(time.RFC3339)
		res.Username = r.cred.Username
		fmt.Println("Authenticated")
	} else {
		session.Options.MaxAge = -1
	}
	res.Session = session
	res.Httpreq = r.httpreq
	return res, err
}

func (s *sessionService) logout(ctx context.Context, r LogoutRequest) (LogoutResponse, error) {
	fmt.Println("Logout service called")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var res LogoutResponse
	session, err := s.store.Get(r.httpreq, "contiv-session")

	if err != nil {
		fmt.Println("error while retrieving session info")
		return LogoutResponse{}, err
	}

	session.Options.MaxAge = -1
	res.Session = session
	res.Httpreq =r.httpreq

	return res, err
}

func (s *sessionService) validateapp(ctx context.Context, r validateAppRequest) (LoginResponse, error) {
	fmt.Println("validate app request service called")
	s.mtx.Lock()
	defer s.mtx.Unlock()
	var res LoginResponse
	session, err := s.store.Get(r.httpreq, "contiv-session")
	if err != nil {
		fmt.Println("error while retrieving session info")
		return LoginResponse{}, err
	}
	if (session.IsNew) {
		fmt.Println("session is new")
		res.Authenticated = false
		res.Message = "Invalid Session validateapp"
		session.Options.MaxAge = -1
	} else {
		fmt.Println("session is present")
		res, err = validate(session)
		if res.Authenticated {
			res.Username = session.Values["Username"].(string)
			session.Values["LastLoginTime"] = time.Now().Format(time.RFC3339)
		}
	}

	res.Session = session
	res.Httpreq = r.httpreq

	return res, err
}

func (s *sessionService) apiprocess(ctx context.Context, r apiRequest) (interface{}, error) {
	fmt.Println("apiget request handler")
	s.mtx.Lock()
	defer s.mtx.Unlock()

	var apiresult apiresponse
	session, err := s.store.Get(r.httpreq, "contiv-session")
	if err != nil {
		fmt.Println("error while retrieving session")
		return apiresult, err
	}

	if session.IsNew {
		fmt.Println("api-get session is new")
		apiresult.sessresponse.Authenticated = false

	} else {
		apiresult.sessresponse, err = validate(session)
		if apiresult.sessresponse.Authenticated {
			fmt.Println("api process session is valid")
			apiresult.result, err = apiexecute(s.apiconfig, r)
		}

	}
	fmt.Println("The session max age = ", session.Options.MaxAge)
	apiresult.sessresponse.Httpreq = r.httpreq
	apiresult.sessresponse.Session = session
	return apiresult, err
}

func apiexecute(apiconfig *apiConfig, r apiRequest) (interface{}, error) {

	var result interface{}
	var err error

	config, ok := validateapi(apiconfig, r)

	if ok {
		switch r.httpreq.Method {

		case "GET": 	fmt.Println("The remote get call =", config.Destination + r.httpreq.URL.Path)
				result, err = httpGet(config.Destination + r.httpreq.URL.Path)
				return result, err
		case "POST":	fmt.Println("The remote post call=", config.Destination + r.httpreq.URL.Path)
				fmt.Println("r.data before post call=", r.data)
				result, err = httpPost(config.Destination + r.httpreq.URL.Path, r.data)
				return result, err
		case "PUT":	fmt.Println("The remote Put call=", config.Destination + r.httpreq.URL.Path)
				result, err = httpPut(config.Destination + r.httpreq.URL.Path, r.data)
				return result, err
		case "DELETE":	fmt.Println("The remote delete call =", config.Destination + r.httpreq.URL.Path)
				err = httpDelete(config.Destination + r.httpreq.URL.Path)
				return result, err

		}
	}
	return result, ErrNotFound
}

func validateapi(apiconfig *apiConfig, r apiRequest) (routedetail, bool) {
	for _, element := range (apiconfig.routelist) {
		if(strings.Contains(r.httpreq.URL.Path, element.Api)){
			if(contains(element.Methods, r.httpreq.Method) >= 0){
				return element, true
			}
		}
	}
	return routedetail{}, false
}

func contains(list interface{}, item interface{}) int {
	for index, element := range list.([]string) {
		if element == item {
			return index
		}
	}
	return -1
}

func httpGet(url string) (interface{}, error){

	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer r.Body.Close()

	switch {
	case r.StatusCode == int(404):
		return nil, errors.New("Page not found!")
	case r.StatusCode == int(403):
		return nil, errors.New("Access denied!")
	case r.StatusCode == int(500):
		response, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(response))
	case r.StatusCode != int(200):
		return nil, errors.New(r.Status)
	}

	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	return response, nil
	//return r, nil
}

func httpPut(url string, jdata interface{}) (interface{}, error) {
	buf, err := json.Marshal(jdata)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(buf)
	req, err := http.NewRequest("PUT", url, body)

	r, err := http.DefaultClient.Do(req)

	defer r.Body.Close()

	switch {
	case r.StatusCode == int(404):
		return nil, errors.New("Page not found!")
	case r.StatusCode == int(403):
		return nil, errors.New("Access denied!")
	case r.StatusCode == int(500):
		response, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(response))

	case r.StatusCode != int(200):
		return nil, errors.New(r.Status)
	}

	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}
	return response, nil
}


func httpPost(url string, jdata interface{}) (interface{}, error) {
	buf, err := json.Marshal(jdata)
	if err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(buf)
	r, err := http.Post(url, "application/json", body)
	if err != nil {
		fmt.Println("the error is =", err.Error())
		return nil, err
	}
	defer r.Body.Close()

	switch {
	case r.StatusCode == int(404):
		return nil, errors.New("Page not found!")
	case r.StatusCode == int(403):
		return nil, errors.New("Access denied!")
	case r.StatusCode == int(500):
		response, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(response))

	case r.StatusCode != int(200):
		return nil, errors.New(r.Status)
	}


	response, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}
	return response, nil
}

func httpDelete(url string) error {

	req, err := http.NewRequest("DELETE", url, nil)

	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	switch {
	case r.StatusCode == int(404):
		return errors.New("Page not found!")
	case r.StatusCode == int(403):
		return errors.New("Access denied!")
	case r.StatusCode == int(500):
		response, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		return errors.New(string(response))

	case r.StatusCode != int(200):
		return errors.New(r.Status)
	}

	return nil
}


func validate(session *sessions.Session) (LoginResponse, error) {
	fmt.Println("validate session called")
	lastLoginTime := session.Values["LastLoginTime"].(string)

	parsedTime, err := time.Parse(time.RFC3339, lastLoginTime)
	if err != nil {
		fmt.Println("Error while parsing time")
		return LoginResponse{}, err
	}
	duration := time.Since(parsedTime)
	fmt.Println("Duration eloped = ", duration.Minutes())
	minutesPassed := duration.Minutes()
	if minutesPassed < 0 || minutesPassed > SessionTimeOut {
		fmt.Println("Session Invalid")
		session.Options.MaxAge = -1
		return LoginResponse{Authenticated: false, Message: "Invalid Session"}, nil
	}
	fmt.Println("Session Valid")
	return LoginResponse{Authenticated: true, Message: "Success"}, nil
}

// Au
