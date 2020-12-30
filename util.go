package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"

	//"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/flynn/json5"
	"github.com/go-errors/errors"
)

// Util.go is a collection of utility functions that are used throughout
// the program

// Count returns the length of a string in runes
// This is exactly equivalent to utf8.RuneCountInString(), just less characters
func Count(s string) int {
	return utf8.RuneCountInString(s)
}

// Convert byte array to rune array
func toRunes(b []byte) []rune {
	runes := make([]rune, 0, utf8.RuneCount(b))

	for len(b) > 0 {
		r, size := utf8.DecodeRune(b)
		runes = append(runes, r)

		b = b[size:]
	}

	return runes
}

func sliceStart(slc []byte, index int) []byte {
	len := len(slc)
	i := 0
	totalSize := 0
	for totalSize < len {
		if i >= index {
			return slc[totalSize:]
		}

		_, size := utf8.DecodeRune(slc[totalSize:])
		totalSize += size
		i++
	}

	return slc[totalSize:]
}

func sliceEnd(slc []byte, index int) []byte {
	len := len(slc)
	i := 0
	totalSize := 0
	for totalSize < len {
		if i >= index {
			return slc[:totalSize]
		}

		_, size := utf8.DecodeRune(slc[totalSize:])
		totalSize += size
		i++
	}

	return slc[:totalSize]
}

// NumOccurrences counts the number of occurrences of a byte in a string
func NumOccurrences(s string, c byte) int {
	var n int
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

// Spaces returns a string with n spaces
func Spaces(n int) string {
	return strings.Repeat(" ", n)
}

// Min takes the min of two ints
func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

// Max takes the max of two ints
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FSize gets the size of a file
func FSize(f *os.File) int64 {
	fi, _ := f.Stat()
	// get the size
	return fi.Size()
}

// IsWordChar returns whether or not the string is a 'word character'
// If it is a unicode character, then it does not match
// Word characters are defined as [A-Za-z0-9_]
func IsWordChar(str string) bool {
	re := regexp.MustCompile(`[\p{L}\p{N}_]`)
	return re.MatchString(str)
}

// IsWhitespace returns true if the given rune is a space, tab, or newline
func IsWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

// IsStrWhitespace returns true if the given string is all whitespace
func IsStrWhitespace(str string) bool {
	for _, c := range str {
		if !IsWhitespace(c) {
			return false
		}
	}
	return true
}

func noAutoCloseChar(str string) bool {
	var x bool

	x = IsWordChar(str)
	if x {
		return x
	}
	if str == `"` || str == "'" || str == "`" {
		return true
	}
	return false
}

// Contains returns whether or not a string array contains a given string
func Contains(list []string, a string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Insert makes a simple insert into a string at the given position
func Insert(str string, pos int, value string) string {
	return string([]rune(str)[:pos]) + value + string([]rune(str)[pos:])
}

// MakeRelative will attempt to make a relative path between path and base
func MakeRelative(path, base string) (string, error) {
	if len(path) > 0 {
		rel, err := filepath.Rel(base, path)
		if err != nil {
			return path, err
		}
		return rel, nil
	}
	return path, nil
}

// BracePairsAreBalanced check if the expression has ballanced brackets
// 0 = Line has balanced brace or dirty close brace " text} , text)" treated as balanced line
// > 0 unbalanced indent
// == -1 balanced } .. {, but requires indent on next line
// == -2 unbalanced but is a closing brace and is the first one
func BracePairsAreBalanced(str string) int {
	// Make string easy to work on with
	// Remove quoted strings so they do not fool the algorithm
	r := regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	str = r.ReplaceAllString(str, "")
	r = regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)
	str = r.ReplaceAllString(str, "")
	// Remove scaped characterd so they do not fool the algorithm
	r = regexp.MustCompile(`\\.`)
	str = r.ReplaceAllString(str, "")
	// Reduce string size
	r = regexp.MustCompile(`^[ \t]+|[ \t]+$`)
	str = r.ReplaceAllString(str, "")
	// Remove {..} [..] from string so it does not fool the algorithm
	str = RemoveNestedBrace(str)
	//messenger.AddLog(str)
	k := 0
	b := 0
	pc := ""
	bs := false
	w := false
	f := false
	for i := 0; i < len(str); i++ {
		c := str[i : i+1]
		if c == "{" || c == "[" || c == "(" {
			b++
			if f == false {
				f = true
				if w == true {
					w = false
				}
			}
		} else if c == "}" || c == "]" || c == ")" {
			if bs == true && (pc == "}" || pc == "]" || pc == ")") {
				if b == -1 {
					b--
				}
			} else {
				b--
				if k == 0 {
					f = true
					bs = true
					b--
				} else if w {
					b++
				}
			}
		} else if k == 0 {
			w = true
		}
		pc = c
		k++
	}
	return b
}

// RemoveNestedBrace Remove nested {..} [..]
func RemoveNestedBrace(str string) string {
	var stack []rune
	var char rune
	var rClose rune
	var rBreak rune

	count := 0
	b := []rune(str)
	for _, c := range b {
		if c == '{' || c == '[' {
			count++
			stack = append(stack, c)
			rBreak = c
			if c == '{' {
				rClose = '}'
			} else {
				rClose = ']'
			}
		} else if c == rClose {
			if count == 0 {
				stack = append(stack, c)
			} else {
				for {
					char, stack = stack[len(stack)-1], stack[:len(stack)-1]
					if char == rBreak || len(stack) <= 0 {
						break
					}
				}
				count--
			}
		} else {
			stack = append(stack, c)
		}
	}
	return string(stack)
}

// BalanceBracePairs Find y line has a balanced braces
func BalanceBracePairs(str string) string {
	n := BracePairsAreBalanced(str)
	if n > 0 || n == -1 {
		return "\t"
	}
	return ""
}

// GetIndentation returns how many leves of indentation in a line depending in the
// indentation char and quantity of characters that make the indentation
// 0 if the line is not indented or -1 if or does not match the indentation char or
// the line has mixed white spaces
func GetLineIndentetion(str, char string, q int) int {
	ws := 0
	crune := []rune(char)

	for _, c := range str {
		if c == ' ' || c == '\t' {
			if crune[0] != c {
				return -1
			}
			ws++
		} else {
			break
		}
	}
	return ws / q
}

// CountLeadingWhitespace returns the ammount of leading whitespace of the given string
func CountLeadingWhitespace(str string) int {
	ws := 0
	for _, c := range str {
		if c == ' ' || c == '\t' {
			ws++
		} else {
			break
		}
	}
	return ws
}

// GetLeadingWhitespace returns the leading whitespace of the given string
func GetLeadingWhitespace(str string) string {
	ws := ""
	for _, c := range str {
		if c == ' ' || c == '\t' {
			ws += string(c)
		} else {
			break
		}
	}
	return ws
}

// IsSpaces checks if a given string is only spaces
func IsSpaces(str []byte) bool {
	for _, c := range str {
		if c != ' ' {
			return false
		}
	}

	return true
}

// IsSpacesOrTabs checks if a given string contains only spaces and tabs
func IsSpacesOrTabs(str string) bool {
	for _, c := range str {
		if c != ' ' && c != '\t' {
			return false
		}
	}

	return true
}

// ParseBool is almost exactly like strconv.ParseBool, except it also accepts 'on' and 'off'
// as 'true' and 'false' respectively
func ParseBool(str string) (bool, error) {
	if str == "on" {
		return true, nil
	}
	if str == "off" {
		return false, nil
	}
	return strconv.ParseBool(str)
}

// EscapePath replaces every path separator in a given path with a %
func EscapePath(path string) string {
	path = filepath.ToSlash(path)
	return strings.Replace(path, "/", "%", -1)
}

// GetModTime returns the last modification time for a given file
// It also returns a boolean if there was a problem accessing the file
func GetModTime(path string) (time.Time, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now(), false
	}
	return info.ModTime(), true
}

// StringWidth returns the width of a string where tabs count as `tabsize` width
func StringWidth(str string, tabsize int) int {
	ns := strings.Count(str, "\t")
	c := Count(str) - ns + ns*tabsize
	return c
}

// RunePos returns the rune index of a given byte index
// This could cause problems if the byte index is between code points
func runePos(p int, str string) int {
	return utf8.RuneCountInString(str[:p])
}

func lcs(a, b string) string {
	arunes := []rune(a)
	brunes := []rune(b)

	lcs := ""
	for i, r := range arunes {
		if i >= len(brunes) {
			break
		}
		if r == brunes[i] {
			lcs += string(r)
		} else {
			break
		}
	}
	return lcs
}

// CommonSubstring gets a common substring among the inputs
func CommonSubstring(arr ...string) string {
	commonStr := arr[0]

	for _, str := range arr[1:] {
		commonStr = lcs(commonStr, str)
	}

	return commonStr
}

// Abs is a simple absolute value function for ints
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// FuncName returns the full name of a given function object
func FuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// ShortFuncName returns the name only of a given function object
func ShortFuncName(i interface{}) string {
	return strings.TrimPrefix(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), "main.(*View).")
}

// ReplaceHome takes a path as input and replaces ~ at the start of the path with the user's
// home directory. Does nothing if the path does not start with '~'.
func ReplaceHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	var userData *user.User
	var err error

	homeString := strings.Split(path, "/")[0]
	if homeString == "~" {
		userData, err = user.Current()
		if err != nil {
			messenger.Error("Could not find user: ", err)
		}
	} else {
		userData, err = user.Lookup(homeString[1:])
		if err != nil {
			if messenger != nil {
				messenger.Error("Could not find user: ", err)
			} else {
				TermMessage("Could not find user: ", err)
			}
			return ""
		}
	}

	home := userData.HomeDir

	return strings.Replace(path, homeString, home, 1)
}

// GetPathAndCursorPosition returns a filename without everything following a `:`
// This is used for opening files like util.go:10:5 to specify a line and column
// Special cases like Windows Absolute path (C:\myfile.txt:10:5) are handled correctly.
func GetPathAndCursorPosition(path string) (string, []string) {
	re := regexp.MustCompile(`([\s\S]+?)(?::(\d+))(?::(\d+))?`)
	match := re.FindStringSubmatch(path)
	// no lines/columns were specified in the path, return just the path with no cursor location
	if len(match) == 0 {
		return path, nil
	} else if match[len(match)-1] != "" {
		// if the last capture group match isn't empty then both line and column were provided
		return match[1], match[2:]
	}
	// if it was empty, then only a line was provided, so default to column 0
	return match[1], []string{match[2], "0"}
}

// ParseCursorLocation get location from cursor position
func ParseCursorLocation(cursorPositions []string) (Loc, error) {
	startpos := Loc{0, 0}
	var err error

	// if no positions are available exit early
	if cursorPositions == nil {
		return startpos, errors.New("No cursor positions were provided")
	}

	startpos.Y, err = strconv.Atoi(cursorPositions[0])
	if err != nil {
		messenger.Error("Error parsing cursor position: ", err)
	} else {
		if len(cursorPositions) > 1 {
			startpos.X, err = strconv.Atoi(cursorPositions[1])
			if err != nil {
				messenger.Error("Error parsing cursor position: ", err)
			}
		}
	}

	return startpos, err
}

// SubstringSafe check no out of bound indices
func SubstringSafe(utf8s string, from int, to int) string {
	lastIndex := Count(utf8s)
	if from > lastIndex {
		return ""
	}
	if to > lastIndex {
		to = lastIndex
	}
	return utf8s[from:to]
}

// DownLoadExtractZip download zip from url and unzip in target dir
func DownLoadExtractZip(url, targetDir string) error {
	//TermMessage(fmt.Sprintf("Downloading %q to %q", url, targetDir))
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = ExtractZip(&data, targetDir)
	if err != nil {
		return err
	}
	return nil
}

// ExtractZip a zip file into a target dir
func ExtractZip(data *[]byte, targetDir string) error {
	zipbuf := bytes.NewReader(*data)
	z, err := zip.NewReader(zipbuf, zipbuf.Size())
	if err != nil {
		return err
	}
	dirPerm := os.FileMode(0750)
	//	if err = os.MkdirAll(targetDir, dirPerm); err != nil {
	//		return err
	//	}

	// Check if all files in zip are in the same directory.
	// this might be the case if the plugin zip contains the whole plugin dir
	// instead of its content.
	var prefix string
	allPrefixed := false
	for i, f := range z.File {
		parts := strings.Split(f.Name, "/")
		if i == 0 {
			prefix = parts[0]
		} else if parts[0] != prefix {
			allPrefixed = false
			break
		} else {
			// switch to true since we have at least a second file
			allPrefixed = true
		}
	}
	// Install files and directory's
	for _, f := range z.File {
		parts := strings.Split(f.Name, "/")
		if allPrefixed {
			parts = parts[1:]
		}

		targetName := filepath.Join(targetDir, filepath.Join(parts...))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(targetName, dirPerm); err != nil {
				return err
			}
		} else {
			basepath := filepath.Dir(targetName)

			if err := os.MkdirAll(basepath, dirPerm); err != nil {
				return err
			}

			content, err := f.Open()
			if err != nil {
				return err
			}
			defer content.Close()
			target, err := os.Create(targetName)
			if err != nil {
				if strings.Contains(err.Error(), "permission denied") {
					// Skip files with permission denied
					continue
				}
				return err
			}
			defer target.Close()
			if _, err = io.Copy(target, content); err != nil {
				return err
			}
		}
	}
	return nil
}

// GeTFileListFromPath gets a full file list from a path
func GeTFileListFromPath(path, extension string) []string {
	Files := []string{}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}
	for _, f := range files {
		if strings.Contains(f.Name(), extension) {
			Files = append(Files, strings.Replace(f.Name(), "."+extension, "", 1))
		}
	}
	return Files
}

// UpdateFileJSON Open an existint JSON file and update existing values and add new ones
// Does not remove existing values not updated
func UpdateFileJSON(filename string, values map[string]string) error {
	// Read JSON FILE, ignore open errors
	ovalues, _ := ReadFileJSON(filename)
	svalues := make(map[string]string)
	// Transform into strings, because we need strings for WriteJSON and we received strings
	for k, v := range ovalues {
		kind := reflect.TypeOf(v).Kind()
		if kind == reflect.Bool {
			svalues[k] = strconv.FormatBool(v.(bool))
		} else if kind == reflect.String {
			svalues[k] = v.(string)
		} else if kind == reflect.Float64 {
			svalues[k] = strconv.FormatFloat(v.(float64), 'f', -1, 64)
		}
	}
	// Now Merge
	for k, v := range values {
		svalues[k] = v
	}
	// Save merged values to JSON FILE
	err := WriteFileJSON(filename, svalues, true)
	if err != nil {
		return err
	}
	return nil
}

// WriteFileJSON helper file to write simple JSON files from a map of strings
func WriteFileJSON(filename string, values map[string]string, parsedValues bool) error {
	var txt []byte

	if parsedValues {
		parsed := make(map[string]interface{})
		for k, v := range values {
			if v == "true" || v == "false" || v == "on" || v == "off" {
				vb, err := strconv.ParseBool(v)
				if err == nil {
					parsed[k] = vb
					continue
				}
			}
			vf, err := strconv.ParseFloat(v, 64)
			if err == nil {
				parsed[k] = vf
				continue
			}
			parsed[k] = v
		}
		txt, _ = json.MarshalIndent(parsed, "", "    ")
	} else {
		txt, _ = json.MarshalIndent(values, "", "    ")
	}
	err := ioutil.WriteFile(filename, append(txt, '\n'), 0644)
	if err != nil {
		return errors.New("Could not write " + filename)
	}
	return nil
}

// ReadFileJSON read a JSON file into a map
func ReadFileJSON(filename string) (map[string]interface{}, error) {
	var parsed map[string]interface{}

	if _, e := os.Stat(filename); e == nil {
		input, err := ioutil.ReadFile(filename)
		if err != nil {
			return parsed, err
		}
		err = json5.Unmarshal(input, &parsed)
		if err != nil {
			return parsed, err
		}
	} else {
		return parsed, errors.New("File does not exists")
	}
	return parsed, nil
}

// Micro-Ide Services

// UnPackSettingFromDownload unzip Settings from download into settings location
func UnPackSettingFromDownload(zipfile *string) error {
	data := []byte(*zipfile)
	err := ExtractZip(&data, configDir)
	if err != nil {
		return err
	}
	return nil
}

// PackSettingsForUpload pack all settings into a zip file for uploading
func PackSettingsForUpload() (string, error) {
	path := strings.ReplaceAll(configDir, "/mi-ide", "") + "/settings.zip"
	err := zipit(configDir, path)
	if err != nil {
		messenger.Error(err.Error())
		return "", err
	}
	zipstring := Slurp(path)
	os.Remove(path)
	return zipstring, nil
}

func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// Slurp a file completely into a string
func Slurp(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return ""
	}
	return string(b)
}
