package excerpt

import (
	"bufio"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineString(t *testing.T) {
	f, _ := os.Open("./source.sqrt")
	str, _ := ioutil.ReadAll(f)
	expected := `a = 1

func test(param)
  print(param)
  ^^^^^^^^^^^^
end

a = 42`
	actual := String(string(str), [4]int{6, 3, 6, 14})
	assert.Equal(t, expected, actual)
}

func TestBlockString(t *testing.T) {
	f, _ := os.Open("./source.sqrt")
	str, _ := ioutil.ReadAll(f)
	expected := `
  a = 1

->func test(param)
->  print(param)
->end

  a = 42
`

	actual := String(string(str), [4]int{5, 3, 7, 14})
	assert.Equal(t, expected, actual)
}

func TestLineFile(t *testing.T) {
	expected := `3  a = 1
4
5  func test(param)
6    print(param)
     ^^^^^^^^^^^^
7  end
8
9  a = 42`
	actual := File("./source.sqrt", [4]int{6, 3, 6, 14})
	assert.Equal(t, expected, actual)
}

func TestBlockFile(t *testing.T) {
	expected := ` 2
 3    a = 1
 4
 5  ->func test(param)
 6  ->  print(param)
 7  ->end
 8
 9    a = 42
10`

	actual := File("./source.sqrt", [4]int{5, 3, 7, 14})
	assert.Equal(t, expected, actual)
}

func TestSnippetLine(t *testing.T) {
	loc := [4]int{6, 3, 6, 14}

	f, _ := os.Open("./source.sqrt")
	str, _ := ioutil.ReadAll(f)
	stringreader := strings.NewReader(string(str))
	f, _ = os.Open("./source.sqrt")
	filereader := bufio.NewReader(f)

	strcode, strinfo := snippet(stringreader, loc)
	filecode, fileinfo := snippet(filereader, loc)

	assert.Equal(t, strcode, filecode)
	assert.Equal(t, strinfo, fileinfo)
	expected := []string{"a = 1", "", "func test(param)", "  print(param)", "end", "", "a = 42"}
	assert.Equal(t, expected, filecode)
	assert.Equal(t, []int{3, 9, 3, 3, 3, 14}, fileinfo)
}

func TestHighlightLine(t *testing.T) {
	input := []string{"a = 1", "", "func test(param)", "  print(param)", "end", "", "a = 42"}
	info := []int{3, 9, 3, 3, 3, 14}

	code, skip := highlightLocation(input, info)

	expected := []string{"a = 1", "", "func test(param)", "  print(param)", "  ^^^^^^^^^^^^", "end", "", "a = 42"}
	assert.Equal(t, 4, skip)
	assert.Equal(t, expected, code)
}

func TestSnippetBlockFile(t *testing.T) {
	loc := [4]int{5, 3, 7, 14}

	f, _ := os.Open("./source.sqrt")
	str, _ := ioutil.ReadAll(f)
	stringreader := strings.NewReader(string(str))
	f, _ = os.Open("./source.sqrt")
	filereader := bufio.NewReader(f)

	strcode, strinfo := snippet(stringreader, loc)
	filecode, fileinfo := snippet(filereader, loc)

	assert.Equal(t, strcode, filecode)
	assert.Equal(t, strinfo, fileinfo)
	expected := []string{"", "a = 1", "", "func test(param)", "  print(param)", "end", "", "a = 42", ""}
	assert.Equal(t, expected, filecode)
	assert.Equal(t, []int{2, 10, 3, 3, 5, 14}, fileinfo)
}

func TestHighlightBlock(t *testing.T) {
	input := []string{"", "a = 1", "", "func test(param)", "  print(param)", "end", "", "a = 42", ""}
	info := []int{2, 10, 3, 3, 5, 14}
	code, skip := highlightLocation(input, info)

	expected := []string{"", "  a = 1", "", "->func test(param)", "->  print(param)", "->end", "", "  a = 42", ""}
	assert.Equal(t, -1, skip)
	assert.Equal(t, expected, code)
}

func TestLineNumbersLine(t *testing.T) {
	input := []string{"a = 1", "", "func test(param)", "  print(param)", "  ^^^^^^^^^^^^", "end", "", "a = 42"}
	info := []int{3, 9, 3, 3, 3, 14}
	skip := 4

	code := lineNumbers(input, info, skip)

	expected := []string{"3  a = 1", "4", "5  func test(param)", "6    print(param)", "     ^^^^^^^^^^^^", "7  end", "8", "9  a = 42"}
	assert.Equal(t, expected, code)
}

func TestLineNumbersBlock(t *testing.T) {
	input := []string{"", "  a = 1", "", "->func test(param)", "->  print(param)", "->end", "", "  a = 42", ""}
	info := []int{2, 10, 3, 3, 5, 14}
	skip := -1

	code := lineNumbers(input, info, skip)

	expected := []string{" 2", " 3    a = 1", " 4", " 5  ->func test(param)", " 6  ->  print(param)", " 7  ->end", " 8", " 9    a = 42", "10"}
	assert.Equal(t, expected, code)
}
