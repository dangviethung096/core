package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type FileHandler func(ctx *HttpContext, filePath string) (HttpResponse, HttpError)

type UploadFileHandler struct {
	URL     Url
	Method  string
	handler func(writer http.ResponseWriter, request *http.Request)
}

func ResgisterFileUpload(url string, method string, handler FileHandler, middlewares ...ApiMiddleware) {
	h := func(writer http.ResponseWriter, request *http.Request) {
		// Create a new context
		ctx := getHttpContext()
		defer putHttpContext(ctx)
		buildContext(ctx, writer, request)

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
		err := request.ParseMultipartForm(10 << 20) // 10 MB
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

		// Create a new file in the server
		dst, err := os.Create(filepath.Join("uploads", fileHeader.Filename))
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
		res, httpErr := handler(ctx, fmt.Sprintf("uploads/%s", fileHeader.Filename))
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
