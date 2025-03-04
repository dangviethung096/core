package core

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type pageInfo struct {
	middleware   []PageMiddleware
	url          string
	handler      PageHandler
	pageFiles    []string
	data         any
	cache        bool
	templateName string
	functionMap  map[string]any
}

type PageRequest struct {
	RoleID    int64
	AccountID int64
	Username  string
	Children  []int64
}

type PageResponse struct {
	PageFiles    []string
	Data         any
	Cache        bool
	TemplateName string
	FunctionMap  map[string]any
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
	ctx.LogInfo("Handle page: %s, requestID = %s", r.URL.String(), ctx.requestID)

	ctx.request = r
	ctx.rw = w
	ctx.URL = r.URL
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

	ctx.LogInfo("Execute handler of request %s, requestID = %s", pageInfo.url, ctx.requestID)
	response, err := pageInfo.handler(ctx, &request)
	if err != nil {
		ctx.LogError("Error when execute handler of request %s: %s", pageInfo.url, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ctx.isResponseEnd {
		return
	}

	pageInfo.pageFiles = response.PageFiles
	pageInfo.data = response.Data
	pageInfo.cache = response.Cache
	pageInfo.templateName = response.TemplateName
	pageInfo.functionMap = response.FunctionMap

	// Render page
	var tmpl *template.Template

	if Config.Server.CacheHtml && pageInfo.cache && htmlTemplateMap[pageInfo.url] != nil {
		tmpl = htmlTemplateMap[pageInfo.url]
	} else {
		tmpl = parseTemplateFile(pageInfo)
		// Set for cache
		htmlTemplateMap[pageInfo.url] = tmpl
	}

	w.Header().Set("Request-ID", ctx.requestID)

	// Execute template
	if originError := tmpl.Execute(w, pageInfo.data); originError != nil {
		ctx.LogError("Error when execute template: %s", originError)
		http.Error(w, originError.Error(), http.StatusInternalServerError)
		return
	}

	ctx.LogInfo("Render page successfully: %s, requestID = %s", pageInfo.url, ctx.requestID)
}

func parseTemplateFile(pageInfo pageInfo) *template.Template {
	pageFiles := pageInfo.pageFiles
	newPageFiles := []string{}
	for _, filePath := range pageFiles {
		if strings.HasSuffix(filePath, "/*") {
			filePath := strings.TrimSuffix(filePath, "/*")
			files, err := listFiles(filePath)
			if err != nil {
				panic(err)
			}
			newPageFiles = append(newPageFiles, files...)
		} else {
			newPageFiles = append(newPageFiles, filePath)
		}
	}

	LogInfo("Parse template file: %#v", newPageFiles)

	tmpl := template.New(pageInfo.templateName)
	tmpl.Funcs(basicFunctionMap)
	if pageInfo.functionMap != nil {
		tmpl = tmpl.Funcs(pageInfo.functionMap)
	}

	tmpl, err := tmpl.ParseFiles(newPageFiles...)
	if err != nil {
		panic(err)
	}
	return tmpl
}

func listFiles(folderPath string) ([]string, Error) {
	folder, err := os.ReadDir(folderPath)
	if err != nil {
		coreContext.LogError("Error when read dir %s: %v", folderPath, err)
		return nil, ERROR_SERVER_ERROR
	}

	filePaths := []string{}
	for _, file := range folder {
		if !file.IsDir() {
			filePaths = append(filePaths, filepath.Join(folderPath, file.Name()))
		}
	}

	return filePaths, nil
}
