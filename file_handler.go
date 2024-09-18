package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type FileHandler func(ctx *HttpContext, filePath string) (HttpResponse, HttpError)

type UploadFileHandler struct {
	URL     Url
	Method  string
	handler func(writer http.ResponseWriter, request *http.Request)
}

func RegisterFileUpload(url string, method string, handler FileHandler, middlewares ...ApiMiddleware) {
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		LogFatal("Error creating uploads directory: %v", err)
	}

	h := func(writer http.ResponseWriter, request *http.Request) {
		// Create a new context
		ctx := getHttpContext()
		defer putHttpContext(ctx)

		ctx.rw = writer
		ctx.request = request
		ctx.requestID = ID.GenerateID()
		ctx.URL = request.URL
		ctx.Method = request.Method

		// Append to common middleware
		middlewareList := []ApiMiddleware{}
		middlewareList = append(middlewareList, commonApiMiddlewares...)
		middlewareList = append(middlewareList, middlewares...)

		// Call middleware of function
		for _, middleware := range middlewareList {
			ctx.isRequestEnd = true
			if err := middleware(ctx); ctx.isRequestEnd {
				if err != nil {
					ctx.writeError(err)
				}
				return
			}
		}

		// Parse the multipart form
		err := request.ParseMultipartForm(MAX_UPLOAD_FILE_SIZE) // 50 MB
		if err != nil {
			ctx.LogError("Error parsing form: %v", err)
			ctx.writeError(NewHttpError(http.StatusInternalServerError, http.StatusInternalServerError, err.Error(), nil))
			return
		}

		// Retrieve the file from form data
		file, fileHeader, err := request.FormFile("file")
		if err != nil {
			ctx.LogError("Error retrieving file: %v", err)
			ctx.writeError(NewHttpError(http.StatusInternalServerError, http.StatusInternalServerError, err.Error(), nil))
			return
		}
		defer file.Close()
		// Append time to file name
		now := time.Now()
		fileName := fmt.Sprintf("%s_%s", now.Format("20060102150405"), fileHeader.Filename)

		// Create a new file in the server
		dst, err := os.Create(filepath.Join("uploads", fileName))
		if err != nil {
			http.Error(writer, "Error creating file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the uploaded file to the server
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(writer, "Error saving file", http.StatusInternalServerError)
			return
		}

		ctx.LogInfo("Request upload file: Url = %s, method = %s, header = %#v", ctx.URL, ctx.Method, ctx.request.Header)
		res, httpErr := handler(ctx, fmt.Sprintf("uploads/%s", fileName))
		if httpErr != nil {
			ctx.LogError("Response error: Url = %s, body = %s", ctx.URL, httpErr.Error())
			ctx.writeError(httpErr)
			return
		}

		if res != nil {
			ctx.LogInfo("Response success: Url = %s, body = %#v", ctx.URL, res)
			ctx.writeSuccess(res)
			return
		}
	}

	uploadFileHandlerMap[url] = UploadFileHandler{
		handler: h,
		URL: Url{
			Path: url,
		},
		Method: method,
	}
}
