package core

type staticFolder struct {
	url    string
	prefix string
	path   string
}

/*
* RegisterFolder registers a folder to a url
* @param url string
* @param prefix string
* @param path string
* @return void
* @example RegisterFolder("/static/", "/static/", "./static")
 */
func RegisterFolder(url string, prefix string, path string) {
	LoggerInstance.Info("Register folder: url = %s, prefix = %s, path = %s", url, prefix, path)
	staticFolder := staticFolder{
		url:    url,
		prefix: prefix,
		path:   path,
	}

	staticFolderMap[url] = staticFolder
}
