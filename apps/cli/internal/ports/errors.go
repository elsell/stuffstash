package ports

import "errors"

var ErrNotLoggedIn = errors.New("not logged in; run stuffstash login")

type Error struct{ Category, Message string }

func (e *Error) Error() string               { return e.Message }
func Failure(category, message string) error { return &Error{category, message} }
