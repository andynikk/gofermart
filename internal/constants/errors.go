package constants

import "errors"

var ErrOrderUpload = errors.New("order upload")
var ErrInvalidFormat = errors.New("invalid format")
var ErrLoginBusy = errors.New("login busy")
var ErrErrorServer = errors.New("error server")
var ErrInvalidLoginPassword = errors.New("invalid login password")
var ErrUserNotAuthenticated = errors.New("user not authenticated")
var ErrAccepted = errors.New("accepted")
var ErrUploadedAnotherUser = errors.New("uploaded another user")
var ErrInvalidOrderNumber = errors.New("invalid order number")
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrNoContent = errors.New("no content")
var ErrConflict = errors.New("conflict")
var ErrTooManyRequests = errors.New("too many requests")
