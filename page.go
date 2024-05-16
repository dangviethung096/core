package core

import (
	"html/template"
	"net/http"
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
	RoleID int64
}

type PageResponse struct {
	PageFiles []string
	Data      any
	Cache     bool
}

type PageHandler func(ctx *HttpContext, request *PageRequest) (PageResponse, Error)

func RegisterPage(url string, handler PageHandler, middleware ...PageMiddleware) {
	LogInfo("Register page: url = %s", url)
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
	// Get http context
	ctx := getHttpContext()
	defer putHttpContext(ctx)

	ctx.request = r
	ctx.rw = w
	// Implement common page middleware
	// Check if middleware is not nil
	request := PageRequest{}

	if pageInfo.middleware != nil {
		// Execute middleware
		for _, middleware := range pageInfo.middleware {
			err := middleware(ctx, &request)
			if err != nil {
				ctx.LogError("Error when execute middleware of request %s: %s", pageInfo.url, err)
			}

			if ctx.isResponseEnd {
				return
			}
		}
	}

	response, err := pageInfo.handler(ctx, &request)
	if err != nil {
		ctx.LogError("Error when execute handler of request %s: %s", pageInfo.url, err)
	}

	if ctx.isResponseEnd {
		return
	}

	pageInfo.pageFiles = response.PageFiles
	pageInfo.data = response.Data
	pageInfo.cache = response.Cache

	// Render page
	var tmpl *template.Template

	if Config.Server.CacheHtml && pageInfo.cache && htmlTemplateMap[pageInfo.url] != nil {
		tmpl = htmlTemplateMap[pageInfo.url]
	} else {
		tmpl = parseTemplateFile(pageInfo)
		// Set for cache
		htmlTemplateMap[pageInfo.url] = tmpl
	}

	// Execute template
	if originError := tmpl.Execute(w, pageInfo.data); originError != nil {
		ctx.LogError("Error when execute template: %s", originError)
		http.Error(w, originError.Error(), http.StatusInternalServerError)
	}
}

func parseTemplateFile(pageInfo pageInfo) *template.Template {
	tmpl, err := template.ParseFiles(pageInfo.pageFiles...)
	if err != nil {
		panic(err)
	}
	return tmpl
}
