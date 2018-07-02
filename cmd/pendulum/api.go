package main

import (
	"fmt"
	"os"
	"path"
	"sort"
	"time"

	"encoding/json"
	"github.com/falconandy/pendulum-agilemarkdown/cmd/agilemarkdown"
	"io/ioutil"
	"net/http"
)

type Location struct {
	Type         string `json:"type"`
	Path         string `json:"path"`
	Name         string `json:"name"`
	Dir          string `json:"dir"`
	LastModified string `json:"last_modified"`
}

type ListResponse struct {
	Folder string     `json:"folder"`
	Files  []Location `json:"files"`
}

type ByFolderAndName []Location

func (a ByFolderAndName) Len() int {
	return len(a)
}

func (a ByFolderAndName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByFolderAndName) Less(i, j int) bool {
	if a[i].Type != a[j].Type {
		if a[i].Type == "dir" {
			return true
		}
		return false
	}
	return a[i].Name < a[j].Name
}

type ReadResponse struct {
	Location
	Contents string `json:"contents"`
}

type StoreResponse struct {
	Status string `json:"status"`
	Log    string `json:"log"`
}

type API struct {
	Path     string
	Assets   http.HandlerFunc
	Contents http.HandlerFunc
}

func (api *API) List(locationPath string) ([]Location, error) {
	response := []Location{}
	fullPath := path.Join(api.Path, locationPath)
	info, err := os.Stat(fullPath)
	if err != nil {
		return response, err
	}
	if !info.IsDir() {
		return response, fmt.Errorf("Path doesn't exist: %s", locationPath)
	}
	files, _ := ioutil.ReadDir(fullPath)
	for _, f := range files {
		name := f.Name()
		if name[0:1] == "." {
			continue
		}
		location := Location{
			Type:         "file",
			Path:         path.Join(locationPath, name),
			Name:         name,
			Dir:          locationPath,
			LastModified: f.ModTime().Format(time.UnixDate),
		}
		if !f.Mode().IsRegular() {
			location.Type = "dir"
			location.Path += "/"
		}
		response = append(response, location)
	}
	sort.Sort(ByFolderAndName(response))
	return response, nil
}

func (api *API) Read(filePath string) (ReadResponse, error) {
	response := ReadResponse{
		Location: Location{
			Type:         "file",
			Path:         filePath,
			Name:         path.Base(filePath),
			Dir:          path.Dir(filePath),
			LastModified: time.Now().Format(time.UnixDate),
		},
	}
	fullPath := path.Join(api.Path, filePath)
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return response, nil
		}
		return response, err
	}
	if !info.Mode().IsRegular() {
		return response, fmt.Errorf("Path is not a file: %s", filePath)
	}
	file, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return response, err
	}
	response.LastModified = info.ModTime().Format(time.UnixDate)
	response.Contents = string(file)
	return response, nil
}

func (api *API) Store(filePath, contents, jwtToken string) (StoreResponse, error) {
	response := StoreResponse{
		Status: "OK",
	}
	fullPath := path.Join(api.Path, filePath)
	err := ioutil.WriteFile(fullPath, []byte(contents), 0644)
	if err != nil {
		return response, err
	}
	response.Log, err = agilemarkdown.Sync(api.Path, jwtToken)
	return response, err
}

func (api *API) Error(err error) interface{} {
	response := struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}{}
	response.Error.Message = fmt.Sprintf("%s", err)
	return response
}

func (api *API) ServeJSON(w http.ResponseWriter, r *http.Request, response interface{}) {
	json, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}
