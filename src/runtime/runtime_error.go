package runtime

import (
	"fmt"
	"strings"

	"github.com/tanema/squirt/src/excerpt"
	"github.com/tanema/squirt/src/lang"
)

type RuntimeErr struct {
	isFile     bool
	file       string
	source     lang.Object
	msg        string
	errorClass string
	errInst    *Instance
	stacktrace []string
}

func (err RuntimeErr) Error() string {
	var clip string
	var lineMsg string
	if err.isFile {
		lineMsg = fmt.Sprintf("%v:%v %v", err.file, err.source.Pos[0], err.msg)
		clip = excerpt.File(err.file, err.source.Pos)
	} else {
		lineMsg = fmt.Sprintf("~:%v %v", err.source.Pos[0], err.msg)
		clip = excerpt.String(err.file, err.source.Pos)
	}
	return fmt.Sprintf(`
%v: %v
%v
%v
%v
	 `,
		err.errorClass,
		err.msg,
		clip,
		lineMsg,
		strings.Join(reverseTrace(err.stacktrace), "\n"),
	)
}

func reverseTrace(trace []string) []string {
	for i, j := 0, len(trace)-1; i < j; i, j = i+1, j-1 {
		trace[i], trace[j] = trace[j], trace[i]
	}
	return trace
}
