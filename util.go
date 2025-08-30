package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"

	//"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
	"github.com/hanspr/shellwords"
	"github.com/hanspr/tcell/v2"
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

// Remove everything to the end starting from index
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

func TrimWhiteSpaceBefore(str string) string {
	nstr := ""
	done := false
	for _, c := range str {
		if done || !IsWhitespace(c) {
			nstr = nstr + string(c)
			done = true
			continue
		}
	}
	return nstr
}

func noAutoCloseChar(str string) bool {
	x := IsWordChar(str)
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

// Prepare line to process with smartindent
func SmartIndentPrepareLine(str string) string {
	// Make string easy to work on with
	// Remove quoted strings so they do not fool the algorithm
	r := regexp.MustCompile(`"(?:[^"\\]|\\.)*"`)
	str = r.ReplaceAllString(str, "")
	r = regexp.MustCompile(`'(?:[^'\\]|\\.)*'`)
	str = r.ReplaceAllString(str, "")
	// Remove scaped characters so they do not fool the algorithm
	r = regexp.MustCompile(`\\.`)
	str = r.ReplaceAllString(str, "")
	// Reduce string size
	r = regexp.MustCompile(`^[ \t]+|[ \t]+$`)
	str = r.ReplaceAllString(str, "")
	// Remove {..} [..] from string so it does not fool the algorithm
	str = RemoveNestedBrace(str)
	return str
}

// RemoveNestedBrace Remove nested {..} [..]
func RemoveNestedBrace(str string) string {
	var stack []rune
	var char rune
	var rClose rune
	var rBreak rune

	count := 0
	//b := []rune(str)
	for _, c := range str {
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

// CountLeadingWhitespace returns the amount of leading whitespace of the given string
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
	if len(str) == 0 {
		return false
	}
	for _, c := range str {
		if c != ' ' {
			return false
		}
	}

	return true
}

// IsSpacesOrTabs checks if a given string contains only spaces and tabs
func IsSpacesOrTabs(str string) bool {
	if len(str) == 0 {
		return false
	}
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
func FuncName(i any) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// ShortFuncName returns the name only of a given function object
func ShortFuncName(i any) string {
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
			messenger.Alert("error", "Could not find user: ", err)
		}
	} else {
		userData, err = user.Lookup(homeString[1:])
		if err != nil {
			if messenger != nil {
				messenger.Alert("error", "Could not find user: ", err)
			} else {
				TermMessage("Could not find user: ", err)
			}
			return ""
		}
	}

	home := userData.HomeDir

	return strings.Replace(path, homeString, home, 1)
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
	data, err := io.ReadAll(resp.Body)
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
	files, err := os.ReadDir(path)
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
		parsed := make(map[string]any)
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
	err := os.WriteFile(filename, append(txt, '\n'), 0644)
	if err != nil {
		return errors.New("Could not write " + filename)
	}
	return nil
}

// ReadFileJSON read a JSON file into a map
func ReadFileJSON(filename string) (map[string]any, error) {
	var parsed map[string]any

	if _, e := os.Stat(filename); e == nil {
		input, err := os.ReadFile(filename)
		if err != nil {
			return parsed, err
		}
		err = json5.Unmarshal(input, &parsed)
		if err != nil {
			return parsed, err
		}
	} else {
		return parsed, errors.New("missing")
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
		messenger.Alert("error", err.Error())
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
	b, err := io.ReadAll(file)
	if err != nil {
		return ""
	}
	return string(b)
}

func ReadHeaderBytes(path string) []byte {
	var line []byte
	file, err := os.Open(path)
	if err != nil {
		return line
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		return scanner.Bytes()
	}
	return line
}

const autocloseOpen = "\"'`({["
const autocloseClose = "\"'`)}]"
const autocloseNewLine = ")}]"

// AutoClose
func AutoClose(v *View, e *tcell.EventKey, isSelection bool, cursorSelection [2]Loc) {
	n, f := isAutoCloseOpen(v, e)
	if !f {
		return
	}
	Close := autocloseClose
	if v.Buf.Settings["charsopen"] != nil && len(v.Buf.Settings["charsopen"].(string)) == len(v.Buf.Settings["charsclose"].(string)) {
		Close = Close + v.Buf.Settings["charsclose"].(string)
	}
	if isSelection {
		x := 1
		if cursorSelection[0].Y < cursorSelection[1].Y {
			x = 0
		}
		v.Cursor.GotoLoc(Loc{cursorSelection[1].X + x, cursorSelection[1].Y})
		v.Buf.Insert(v.Cursor.Loc, string([]rune(Close)[n]))
		return
	}
	m := strings.Index(Close, string(e.Rune()))
	if n != -1 || m != -1 {
		if n < 3 || m > 2 {
			// Test we do not duplicate closing char
			x := v.Cursor.X
			if v.Cursor.X < len(v.Buf.LineRunes(v.Cursor.Y)) && v.Cursor.X > 1 {
				chb := string(v.Buf.LineRunes(v.Cursor.Y)[x : x+1])
				if chb == string(e.Rune()) {
					v.Delete(false)
					n = -1
				}
			}
			if n >= 0 && n < 3 && v.Cursor.X < len(v.Buf.LineRunes(v.Cursor.Y))-1 && v.Cursor.X > 1 {
				// Test surrounding chars
				ch1 := string(v.Buf.LineRunes(v.Cursor.Y)[x-2 : x-1])
				ch2 := string(v.Buf.LineRunes(v.Cursor.Y)[x : x+1])
				if noAutoCloseChar(ch1) || noAutoCloseChar(ch2) {
					n = -1
				}
			}
		}
		if n >= 0 {
			v.Buf.Insert(v.Cursor.Loc, string([]rune(Close)[n]))
			v.Cursor.Left()
		}
	}
}

func isAutoCloseOpen(v *View, e *tcell.EventKey) (int, bool) {
	Open := autocloseOpen
	if v.Buf.Settings["charsopen"] != nil && len(v.Buf.Settings["charsopen"].(string)) == len(v.Buf.Settings["charsclose"].(string)) {
		Open = Open + v.Buf.Settings["charsopen"].(string)
	}
	n := -1
	for _, runeValue := range Open {
		n++
		if runeValue == e.Rune() {
			return n, true
		}
	}
	return -1, false
}

// Shell commands

// ExecCommand executes a command using exec
// It returns any output/errors
func ExecCommand(name string, arg ...string) (string, error) {
	var err error
	cmd := exec.Command(name, arg...)
	outputBytes := &bytes.Buffer{}
	cmd.Stdout = outputBytes
	cmd.Stderr = outputBytes
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	err = cmd.Wait() // wait for command to finish
	outstring := outputBytes.String()
	return outstring, err
}

// RunShellCommand executes a shell command and returns the output/error
func RunShellCommand(input string) (string, error) {
	args, err := shellwords.Split(input)
	if err != nil {
		return "", err
	}
	inputCmd := args[0]

	return ExecCommand(inputCmd, args[1:]...)
}

// RunBackgroundShell running shell in background
func RunBackgroundShell(input string) {
	args, err := shellwords.Split(input)
	if err != nil {
		messenger.Alert("error", err)
		return
	}
	inputCmd := args[0]
	messenger.Message("Running...")
	go func() {
		output, err := RunShellCommand(input)
		totalLines := strings.Split(output, "\n")

		if len(totalLines) < 3 {
			if err == nil {
				messenger.Message(inputCmd, " exited without error")
			} else {
				messenger.Message(inputCmd, " exited with error: ", err, ": ", output)
			}
		} else {
			messenger.Message(output)
		}
		// We have to make sure to redraw
		RedrawAll(true)
	}()
}

// GetProjectDir tries to find the correct path to the project
// with respecto to the file directory path
func GetProjectDir(wkdir, fdir string) string {
	// search for existing settings on dir or above
	for _, path := range []string{wkdir, fdir} {
		dirPath := path + "/.miide"
		_, err := os.Stat(dirPath)
		if err == nil {
			return path
		}
		dirPath = filepath.Dir(path)
		_, err = os.Stat(dirPath)
		if err == nil {
			return dirPath
		}

	}
	// could not find local settings, set the best option
	if wkdir == homeDir {
		return fdir
	}
	if wkdir != fdir && strings.Contains(fdir, wkdir) {
		return wkdir
	}
	return fdir
}
