package lang

import (
	"fmt"

	"github.com/tanema/squirt/src/excerpt"
)

type ParseErr struct {
	file   bool
	source string
	msg    string
	token  token
}

func (err ParseErr) Error() string {
	var clip string
	var filename string
	if err.file {
		clip = excerpt.File(err.source, err.token.loc)
		filename = err.source
	} else {
		clip = excerpt.String(err.source, err.token.loc)
		filename = "~"
	}
	return fmt.Sprintf(`Parse Error: %v
%v
%v:%v:%v
	`,
		err.msg,
		clip,
		filename,
		err.token.loc[0],
		err.token.loc[1],
	)
}
