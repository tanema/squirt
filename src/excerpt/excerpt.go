package excerpt

import (
	"bufio"
	"io"
	"os"
	"strconv"
	"strings"
)

const LinePad = 3

func File(filename string, loc [4]int) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	return excerpt(bufio.NewReader(f), loc, true)
}

func String(str string, loc [4]int) string {
	return excerpt(strings.NewReader(str), loc, false)
}

func excerpt(r io.Reader, loc [4]int, lineNums bool) string {
	code, info := snippet(r, loc)
	code, skip := highlightLocation(code, info)
	if lineNums {
		code = lineNumbers(code, info, skip)
	}
	return strings.Join(code, "\n")
}

func snippet(r io.Reader, loc [4]int) ([]string, []int) {
	scanner := bufio.NewScanner(r)
	out := []string{}
	lineStart, _, lineEnd, _ := loc[0], loc[1], loc[2], loc[3]
	start, end := max(1, lineStart-LinePad), lineEnd+LinePad-1
	line := 0
	success := scanner.Scan()
	for success && line <= end {
		line++
		if line >= start {
			out = append(out, scanner.Text())
		}
		success = scanner.Scan()
	}
	return out, []int{start, line, lineStart - start, loc[1], lineEnd - start, loc[3]}
}

func highlightLocation(code []string, loc []int) ([]string, int) {
	lineStart, colStart, lineEnd, colEnd := loc[2], loc[3], loc[4], loc[5]
	if lineStart == lineEnd {
		highlight := leftPad(repeat("^", 1+colEnd-colStart), colEnd)
		code = append(code[:lineStart+1], append([]string{highlight}, code[lineStart+1:]...)...)
		return code, lineStart + 1
	}
	for i, str := range code {
		if i >= lineStart && i <= lineEnd {
			code[i] = "->" + str
		} else if str != "" {
			code[i] = "  " + str
		}
	}
	return code, -1
}

func lineNumbers(code []string, loc []int, skip int) []string {
	max := digits(loc[1])
	for i, str := range code {
		num := i + loc[0]
		if i == skip {
			code[i] = repeat(" ", max+2) + str
			continue
		}
		if i > skip && skip != -1 {
			num--
		}
		if str != "" {
			str = "  " + str
		}
		code[i] = leftPad(strconv.Itoa(num), max) + str
	}
	return code
}

func leftPad(str string, desiredLen int) string {
	left := desiredLen - len(str)
	if left <= 0 {
		return str
	}
	return repeat(" ", left) + str
}

func digits(i int) int {
	count := 0
	for i > 0 {
		i = i / 10
		count++
	}
	return count
}

func repeat(str string, rep int) string {
	if rep <= 0 {
		return ""
	}
	return strings.Repeat(str, rep)
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
