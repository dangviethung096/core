package core

import (
	"html/template"
	"net/http"
)

type Page struct {
	middleware PageMiddleware
	url        string
	PageFiles  []string
	Data       any
}

func RegisterPage(url string, pageInfo Page) {
	coreContext.LogInfo("Register page: url = %s, pageFiles = %#v", url, pageInfo.PageFiles)
	pageInfo.url = url
	if Config.Server.CacheHtml {
		// Parse files html
		tmpl := parseTemplateFile(pageInfo)
		htmlTemplateMap[url] = tmpl
	}

	pageMap[url] = pageInfo
}

func RegisterPageWithMiddleware(url string, pageInfo Page, middleware PageMiddleware) {
	coreContext.LogInfo("Register page: url = %s, pageFiles = %#v", url, pageInfo.PageFiles)
	pageInfo.url = url
	if Config.Server.CacheHtml {
		// Parse files html
		tmpl := parseTemplateFile(pageInfo)
		htmlTemplateMap[url] = tmpl
	}

	pageInfo.middleware = middleware
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
			err := pageInfo.middleware(w, r)
			if err != nil {
				LoggerInstance.Error("Error when execute middleware of request %s: %s", pageInfo.url, err)
				return
			}
		}

		// Render page
		var tmpl *template.Template
		var err error
		if Config.Server.CacheHtml {
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
