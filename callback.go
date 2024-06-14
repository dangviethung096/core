package core

type CallbackFunc func()

var callback map[string]CallbackFunc

func AddCallback(key string, cb CallbackFunc) {
	callback[key] = cb
}
