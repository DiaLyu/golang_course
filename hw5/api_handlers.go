package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type ResponseBody struct {
	Error    string      `json:"error"`
	Response interface{} `json:"response,omitempty"`
}

func (srv *MyApi) handleProfile(w http.ResponseWriter, req *http.Request) {
	params := ProfileParams{}

	loginStr := req.FormValue("login")
	params.Login = loginStr

	if params.Login == "" {
		err := fmt.Errorf("login must me not empty")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	res, err := srv.Profile(req.Context(), params)
	if err != nil {
		if err = respondError(w, err); err != nil {
			log.Println(err)
		}
		return
	}
	
	body := ResponseBody{Response: res}
	payload, err := json.Marshal(body)
	if err != nil {
		if err = respondError(w, err); err != nil {
			log.Println(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		log.Println(err)
	}
}

func (srv *MyApi) handleCreate(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("X-Auth") != "100500" {
		err := fmt.Errorf("unauthorized")
		if err = respondError(w, ApiError{http.StatusForbidden, err}); err != nil {
			log.Println(err)
		}
		return
	}
	
	params := CreateParams{}

	loginStr := req.FormValue("login")
	params.Login = loginStr

	if params.Login == "" {
		err := fmt.Errorf("login must me not empty")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	if len(params.Login) < 10 {
		err := fmt.Errorf("login len must be >= 10")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	nameStr := req.FormValue("full_name")
	params.Name = nameStr

	statusStr := req.FormValue("status")
	params.Status = statusStr

	if params.Status == "" {
		params.Status = "user"
	}

	switch params.Status {
	case "user", "moderator", "admin":
	default:
		err := fmt.Errorf("status must be one of [user, moderator, admin]")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	ageStr := req.FormValue("age")
	ageInt, err := strconv.Atoi(ageStr)
	if err != nil {
		err = fmt.Errorf("age must be int")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
    }
	params.Age = ageInt

	if params.Age < 0 {
		err := fmt.Errorf("age must be >= 0")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	if params.Age > 128 {
		err := fmt.Errorf("age must be <= 128")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	res, err := srv.Create(req.Context(), params)
	if err != nil {
		if err = respondError(w, err); err != nil {
			log.Println(err)
		}
		return
	}
	
	body := ResponseBody{Response: res}
	payload, err := json.Marshal(body)
	if err != nil {
		if err = respondError(w, err); err != nil {
			log.Println(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		log.Println(err)
	}
}

func (srv *OtherApi) handleCreate(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("X-Auth") != "100500" {
		err := fmt.Errorf("unauthorized")
		if err = respondError(w, ApiError{http.StatusForbidden, err}); err != nil {
			log.Println(err)
		}
		return
	}
	
	params := OtherCreateParams{}

	usernameStr := req.FormValue("username")
	params.Username = usernameStr

	if params.Username == "" {
		err := fmt.Errorf("username must me not empty")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	if len(params.Username) < 3 {
		err := fmt.Errorf("username len must be >= 3")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	nameStr := req.FormValue("account_name")
	params.Name = nameStr

	classStr := req.FormValue("class")
	params.Class = classStr

	if params.Class == "" {
		params.Class = "warrior"
	}

	switch params.Class {
	case "warrior", "sorcerer", "rouge":
	default:
		err := fmt.Errorf("class must be one of [warrior, sorcerer, rouge]")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	levelStr := req.FormValue("level")
	levelInt, err := strconv.Atoi(levelStr)
	if err != nil {
		err = fmt.Errorf("level must be int")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
    }
	params.Level = levelInt

	if params.Level < 1 {
		err := fmt.Errorf("level must be >= 1")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	if params.Level > 50 {
		err := fmt.Errorf("level must be <= 50")
		if err = respondError(w, ApiError{http.StatusBadRequest, err}); err != nil {
			log.Println(err)
		}
		return
	}

	res, err := srv.Create(req.Context(), params)
	if err != nil {
		if err = respondError(w, err); err != nil {
			log.Println(err)
		}
		return
	}
	
	body := ResponseBody{Response: res}
	payload, err := json.Marshal(body)
	if err != nil {
		if err = respondError(w, err); err != nil {
			log.Println(err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		log.Println(err)
	}
}


func respondError(w http.ResponseWriter, err error) error {
	if apiErr, ok := err.(ApiError); ok {
		view := ResponseBody{Error: apiErr.Err.Error()}

		payload, err := json.Marshal(view)
		if err != nil {
			return fmt.Errorf("failed to marshal error due %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.HTTPStatus)
		_, err = w.Write(payload)
		if err != nil {
			return fmt.Errorf("failed to write response body due %v", err)
		}

		return nil
	}

	return respondError(w, ApiError{
		HTTPStatus: http.StatusInternalServerError,
		Err:        err,
	})
}


func (srv *MyApi) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	
	case "/user/profile":
		
		srv.handleProfile(w, req)
	
	case "/user/create":
		
		if req.Method != "POST" {
			err := fmt.Errorf("bad method")
			if err = respondError(w, ApiError{http.StatusNotAcceptable, err}); err != nil {
				log.Println(err)
			}
			return
		}
		
		srv.handleCreate(w, req)
	
	default:
		err := fmt.Errorf("unknown method")
		if err = respondError(w, ApiError{http.StatusNotFound, err}); err != nil {
			log.Println(err)
		}
	}
}

func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	
	case "/user/create":
		
		if req.Method != "POST" {
			err := fmt.Errorf("bad method")
			if err = respondError(w, ApiError{http.StatusNotAcceptable, err}); err != nil {
				log.Println(err)
			}
			return
		}
		
		srv.handleCreate(w, req)
	
	default:
		err := fmt.Errorf("unknown method")
		if err = respondError(w, ApiError{http.StatusNotFound, err}); err != nil {
			log.Println(err)
		}
	}
}
