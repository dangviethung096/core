package core

import (
	"html/template"
	"net/http"
	"net/url"
)

type pageInfo struct {
	middleware []PageMiddleware
	url        string
	handler    PageHandler
	pageFiles  []string
	data       any
	cache      bool
}

type PageRequest struct {
	Header  http.Header
	Queries url.Values
}

type PageResponse struct {
	PageFiles []string
	Data      any
	Cache     bool
}

type PageHandler func(url string, request PageRequest) PageResponse

func RegisterPage(url string, handler PageHandler, middleware ...PageMiddleware) {
	coreContext.LogInfo("Register page: url = %s", url)
	pageInfo := pageInfo{
		url:     url,
		handler: handler,
		cache:   true,
	}

	if len(middleware) > 0 {
		pageInfo.middleware = middleware
	} else {
		pageInfo.middleware = nil
	}

	if Config.Server.CacheHtml {
		// Parse files html
		tmpl := parseTemplateFile(pageInfo)
		htmlTemplateMap[url] = tmpl
	}

	pageMap[url] = pageInfo
}

/*
* pageHandler is a handler function that will render page
* It will check if middleware is not nil, execute middleware
* If middleware return error, page will not be rendered
* If middleware return nil, page will be rendered
 */
func pageHandler(pageInfo pageInfo, w http.ResponseWriter, r *http.Request) {
	// Implement common page middleware
	// Check if middleware is not nil
	if pageInfo.middleware != nil {
		// Execute middleware
		for _, middleware := range pageInfo.middleware {
			err := middleware(w, r)
			if err != nil {
				LogError("Error when execute middleware of request %s: %s", pageInfo.url, err)
				return
			}
		}
	}

	request := PageRequest{
		Header:  r.Header,
		Queries: r.URL.Query(),
	}

	response := pageInfo.handler(pageInfo.url, request)
	pageInfo.pageFiles = response.PageFiles
	pageInfo.data = response.Data
	pageInfo.cache = response.Cache

	// Render page
	var tmpl *template.Template
	var err error

	if Config.Server.CacheHtml && pageInfo.cache && htmlTemplateMap[pageInfo.url] != nil {
		tmpl = htmlTemplateMap[pageInfo.url]
	} else {
		tmpl = parseTemplateFile(pageInfo)
		// Set for cache
		htmlTemplateMap[pageInfo.url] = tmpl
	}

	// Execute template
	err = tmpl.Execute(w, pageInfo.data)
	if err != nil {
		LogError("Error when execute template: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func parseTemplateFile(pageInfo pageInfo) *template.Template {
	tmpl, err := template.ParseFiles(pageInfo.pageFiles...)
	if err != nil {
		panic(err)
	}
	return tmpl
}
