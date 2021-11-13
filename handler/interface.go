package handler

import "matrixChannel/config"

type handler interface {
	GetName() string
	New(C *config.Config, userIdx string) handler
	Do(conf *config.Config) error
}
