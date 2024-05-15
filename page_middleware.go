package core

/*
* PageMiddleware is a type of function that takes a
* http.ResponseWriter and a http.Request and returns an Error
* If it return Error, page will not be rendered
* If it return nil, page will be rendered
 */
type PageMiddleware func(*HttpContext, *PageRequest) Error
