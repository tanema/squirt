package _test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tanema/squirt/src/runtime"
)

func TestIO(t *testing.T) {
	filepath.Walk("./examples", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return err
		}
		println(path)
		var actual strings.Builder
		scope := runtime.DefaultNamespace(&actual)
		tbl, _ := runtime.ToValue(scope, []runtime.Value{})
		scope.Set("ARGV", tbl)
		_, err = runtime.EvalFile(scope, path)
		assert.Nil(t, err)
		expected, _ := ioutil.ReadFile(strings.Replace(path, "examples", "outputs", -1))
		assert.Equal(t, string(expected), actual.String(), fmt.Sprintf("mismatch in output of %v", path))
		return nil
	})
}
