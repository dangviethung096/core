package core

import (
	"html/template"
	"net/http"
	"net/url"
)

type Page struct {
	middleware []PageMiddleware
	url        string
	handler    PageHandler
	PageFiles  []string
	Data       any
	Cache      bool
}

type PageRequest struct {
	Header  http.Header
	Queries url.Values
}

type PageResponse struct {
	PageFiles []string
	Data      any
}

type PageHandler func(url string, request PageRequest) PageResponse

func RegisterPage(url string, handler PageHandler) {
	coreContext.LogInfo("Register page: url = %s", url)
	pageInfo := Page{
		url:        url,
		handler:    handler,
		Cache:      true,
		middleware: nil,
	}

	if Config.Server.CacheHtml {
		// Parse files html
		tmpl := parseTemplateFile(pageInfo)
		htmlTemplateMap[url] = tmpl
	}

	pageMap[url] = pageInfo
}

func RegisterPageWithMiddleware(url string, handler PageHandler, middleware PageMiddleware) {
	coreContext.LogInfo("Register page: url = %s", url)
	pageInfo := Page{
		handler:    handler,
		url:        url,
		middleware: []PageMiddleware{middleware},
	}

	pageMap[url] = pageInfo
}

/*
* pageHandler is a handler function that will render page
* It will check if middleware is not nil, execute middleware
* If middleware return error, page will not be rendered
* If middleware return nil, page will be rendered
 */
func pageHandler(pageInfo Page) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if middleware is not nil
		if pageInfo.middleware != nil {
			// Execute middleware
			for _, middleware := range pageInfo.middleware {
				err := middleware(w, r)
				if err != nil {
					LoggerInstance.Error("Error when execute middleware of request %s: %s", pageInfo.url, err)
					return
				}
			}
		}

		request := PageRequest{
			Header:  r.Header,
			Queries: r.URL.Query(),
		}

		response := pageInfo.handler(pageInfo.url, request)
		pageInfo.PageFiles = response.PageFiles
		pageInfo.Data = response.Data

		// Render page
		var tmpl *template.Template
		var err error

		if Config.Server.CacheHtml && pageInfo.Cache && htmlTemplateMap[pageInfo.url] != nil {
			tmpl = htmlTemplateMap[pageInfo.url]
		} else {
			tmpl = parseTemplateFile(pageInfo)
		}

		// Execute template
		err = tmpl.Execute(w, pageInfo.Data)
		if err != nil {
			LoggerInstance.Error("Error when execute template: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func parseTemplateFile(pageInfo Page) *template.Template {
	tmpl, err := template.ParseFiles(pageInfo.PageFiles...)
	if err != nil {
		panic(err)
	}
	return tmpl
}
