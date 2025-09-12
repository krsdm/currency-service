package password

import "errors"

var (
	ErrShortPassword           = errors.New("short password")
	ErrLongPassword            = errors.New("long password")
	ErrProtectedPasswordFormat = errors.New("invalid protected password format")
	ErrInvalidPassword         = errors.New("invalid password")
)

type ErrRequiredSymbols struct {
	msg string
}

func NewErrRequiredSymbols(msg string) ErrRequiredSymbols {
	return ErrRequiredSymbols{msg}
}

func (e ErrRequiredSymbols) Error() string {
	return e.msg
}
