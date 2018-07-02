package main

import (
	"github.com/titpetric/pendulum/cmd/agilemarkdown"
	"net/http"
	"strings"
)

func (api *API) ListHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	response := struct {
		Response ListResponse `json:"response"`
	}{}

	response.Response.Folder = r.URL.Path
	response.Response.Files, err = api.List(strings.Replace(r.URL.Path, "/api/list", "", 1))
	if err != nil {
		api.ServeJSON(w, r, api.Error(err))
		return
	}
	api.ServeJSON(w, r, response)
}

func (api *API) ReadHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	jwtTokenCookie, err := r.Cookie("jwt_token")
	var jwtToken string
	if err == nil && jwtTokenCookie != nil {
		jwtToken = jwtTokenCookie.Value
	}
	agilemarkdown.CreateUserIfNotExist(api.Path, jwtToken)

	response := struct {
		Response ReadResponse `json:"response"`
	}{}
	response.Response, err = api.Read(strings.Replace(r.URL.Path, "/api/read", "", 1))

	if err != nil {
		api.ServeJSON(w, r, api.Error(err))
		return
	}
	api.ServeJSON(w, r, response)
}

func (api *API) StoreHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	jwtTokenCookie, err := r.Cookie("jwt_token")
	var jwtToken string
	if err == nil && jwtTokenCookie != nil {
		jwtToken = jwtTokenCookie.Value
	}

	response := struct {
		Response StoreResponse `json:"response"`
	}{}
	response.Response, err = api.Store(strings.Replace(r.URL.Path, "/api/store", "", 1), r.PostFormValue("contents"), jwtToken)

	if err != nil {
		api.ServeJSON(w, r, api.Error(err))
		return
	}
	api.ServeJSON(w, r, response)
}

func (api *API) AddIdeaHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	ideaTitle := r.FormValue("title")
	jwtTokenCookie, err := r.Cookie("jwt_token")
	var jwtToken string
	if err == nil && jwtTokenCookie != nil {
		jwtToken = jwtTokenCookie.Value
	}

	ideaPath, ideaContent, err := agilemarkdown.AddIdea(api.Path, ideaTitle, jwtToken)
	if err != nil {
		api.ServeJSON(w, r, api.Error(err))
		return
	}

	response := struct {
		Response ReadResponse `json:"response"`
	}{}
	response.Response, err = api.Read(ideaPath)
	if response.Response.Contents == "" {
		response.Response.Contents = ideaContent
	}

	if err != nil {
		api.ServeJSON(w, r, api.Error(err))
		return
	}
	api.ServeJSON(w, r, response)
}

func (api *API) AddStoryHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	storyTitle := r.FormValue("title")
	project := r.FormValue("project")
	jwtTokenCookie, err := r.Cookie("jwt_token")
	var jwtToken string
	if err == nil && jwtTokenCookie != nil {
		jwtToken = jwtTokenCookie.Value
	}

	storyPath, storyContent, err := agilemarkdown.AddStory(api.Path, project, storyTitle, jwtToken)
	if err != nil {
		api.ServeJSON(w, r, api.Error(err))
		return
	}

	response := struct {
		Response ReadResponse `json:"response"`
	}{}
	response.Response, err = api.Read(storyPath)
	if response.Response.Contents == "" {
		response.Response.Contents = storyContent
	}

	if err != nil {
		api.ServeJSON(w, r, api.Error(err))
		return
	}
	api.ServeJSON(w, r, response)
}
