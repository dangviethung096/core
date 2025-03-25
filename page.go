package core

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type pageInfo struct {
	middleware   []PageMiddleware
	url          string
	handler      PageHandler
	pageFiles    []string
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
