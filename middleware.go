package core

import "net/http"

func corsMiddleware(ctx *HttpContext) HttpError {
	ctx.rw.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	ctx.rw.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if ctx.Method == http.MethodOptions {
		ctx.rw.WriteHeader(http.StatusOK)
		return nil
	}

	ctx.Next()
	return nil
}

func UseCorsMiddleware() {
	UseMiddleware(corsMiddleware)
}

func UseMiddleware(middleware ApiMiddleware) {
	commonApiMiddlewares = append(commonApiMiddlewares, middleware)
}
