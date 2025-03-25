package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type FileHandler func(ctx *HttpContext, filePath string) (HttpResponse, HttpError)

func RegisterFileUpload(url string, method string, handler FileHandler, middlewares ...ApiMiddleware) {
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		LogFatal("Error creating uploads directory: %v", err)
	}

	h := func(c *gin.Context) {
		// Create a new context
		ctx := getHttpContext(c)
		defer putHttpContext(ctx)

		ctx.rw = ctx.Writer
		ctx.request = ctx.Request
		ctx.URL = ctx.Request.URL
		ctx.Method = ctx.Request.Method

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
		err := ctx.request.ParseMultipartForm(MAX_UPLOAD_FILE_SIZE) // 50 MB
		if err != nil {
			ctx.LogError("Error parsing form: %v", err)
			ctx.writeError(NewHttpError(http.StatusInternalServerError, http.StatusInternalServerError, err.Error(), nil))
			return
		}

		// Retrieve the file from form data
		file, fileHeader, err := ctx.request.FormFile("file")
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
			http.Error(ctx.Writer, "Error creating file", http.StatusInternalServerError)
			return
		}

		defer func() {
			// Remove file after handle
			err = os.Remove(fmt.Sprintf("uploads/%s", fileName))
			if err != nil {
				ctx.LogError("Error remove file: %v", err)
			}
		}()

		defer dst.Close()

		// Copy the uploaded file to the server
		_, err = io.Copy(dst, file)
		if err != nil {
			http.Error(ctx.Writer, "Error saving file", http.StatusInternalServerError)
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

	switch method {
	case http.MethodPost:
		router.POST(url, h)
	case http.MethodGet:
		router.GET(url, h)
	case http.MethodPut:
		router.PUT(url, h)
	case http.MethodDelete:
		router.DELETE(url, h)
	}
}
