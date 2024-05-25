package mock

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"regexp"
	"strings"
)

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}

func getFile(filepath string) (*os.File, error) {
	if filepath == "" {
		return nil, errors.New("please input a file")
	}

	if existed, err := fileExists(filepath); existed && err != nil {
		return nil, errors.New("the file provided does not exist")
	}

	file, e := os.Open(filepath)

	if e != nil {
		return nil, errors.Wrapf(e, "unable to read the file %s", filepath)
	}
	return file, nil
}

var preMatchReg = regexp.MustCompile("is a mock of .+? interface.")

func patchMock(r io.Reader, w io.Writer) error {
	reader := bufio.NewReader(r)

	var preMatchFlag bool

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		_, err = w.Write(line)
		if err != nil {
			return err
		}

		_, _ = w.Write([]byte("\n"))

		lineStr := string(line)

		if strings.HasPrefix(lineStr, "import (") {
			_, err = w.Write([]byte("\tgoneMock \"github.com/gone-io/gone\"\n"))
			if err != nil {
				return err
			}
		}

		if preMatchReg.Match(line) {
			preMatchFlag = true
		}

		if preMatchFlag && strings.HasPrefix(lineStr, "type") {
			_, err = w.Write([]byte("\tgoneMock.Flag\n"))
			if err != nil {
				return err
			}
			preMatchFlag = false
		}
	}

	scanner := bufio.NewScanner(bufio.NewReader(r))
	for scanner.Scan() {
		_, e := fmt.Fprintln(w, strings.ToUpper(scanner.Text()))
		if e != nil {
			return e
		}
	}
	return nil
}

func fileExists(filepath string) (bool, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return false, nil
	} else {
		return true, nil
	}
}
