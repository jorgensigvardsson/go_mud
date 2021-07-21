package absmachine

const (
	ErrorIconsistency = iota + 1
	ErrorInvalidDirection
)

type LowLevelOpsError struct {
	errorCode int
	message   string
}

func (e *LowLevelOpsError) Error() string {
	return e.message
}

func (e *LowLevelOpsError) ErrorCode() int {
	return e.errorCode
}
