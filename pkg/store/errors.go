package store

import (
	"errors"
)

var ErrNameEmpty = errors.New("name must be set")
var ErrNamespaceEmpty = errors.New("namespace must be set")
var ErrInvalid = errors.New("object is invalid")
var ErrNotFound = errors.New("object not found")
var ErrAlreadyExists = errors.New("object already exists")
var ErrObjectChanged = errors.New("newer object version in store, please get the latest version")
