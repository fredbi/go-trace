package itracer

import (
	"path"
	"runtime"

	"go.uber.org/zap"
)

// SignedFields prepends a slice of zap.Fields with a signature string
func SignedFields(signature string, fields []zap.Field) []zap.Field {
	signedFields := make([]zap.Field, 0, len(fields)+1)
	signedFields = append(signedFields, zap.String(prefix, signature))
	signedFields = append(signedFields, fields...)

	return signedFields
}

// Signature returns the function name of the caller.
func Signature() string {
	pc, _, _, ok := runtime.Caller(2)
	if ok {
		return path.Base(runtime.FuncForPC(pc).Name())
	}

	return ""
}
