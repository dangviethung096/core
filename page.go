package core

import (
	"html/template"
	"net/http"
)

type Page struct {
	url       string
	PageFiles []string
	Data      any
}

func RegisterPage(url string, pageInfo Page) {
	LoggerInstance.Info("Register page: url = %s, pageFiles = %#v", url, pageInfo.PageFiles)
	pageInfo.url = url
	if Config.Server.CacheHtml {
		// Parse files html
		tmpl, err := template.ParseFiles((pageInfo.PageFiles[0]))
		if err != nil {
			LoggerInstance.Error("Parse html file fail: %s", err.Error())
			panic(err)
		}

		htmlTemplateMap[url] = tmpl
	}

	pageMap[url] = pageInfo
}

func pageHandler(pageInfo Page) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tmpl *template.Template
		var err error
		if Config.Server.CacheHtml {
			tmpl = htmlTemplateMap[pageInfo.url]
		} else {
			// Parse files html
			tmpl, err = template.ParseFiles((pageInfo.PageFiles[0]))
			if err != nil {
				panic(err)
			}
		}

		// Execute template
		err = tmpl.Execute(w, pageInfo.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
