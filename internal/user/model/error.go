package model

import "errors"

var ErrNoUser = errors.New("user is not exist")
var ErrUserAlreadyExist = errors.New("user is already existed")
var ErrInvalidParameter = errors.New("invalid paramter")
var ErrInternalServerError = errors.New("something went wrong in server, you can try again")
var ErrNoPermission = errors.New("user does not have permission")
