package mock

import (
	"bufio"
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

	if fileInfo, err := os.Stat(filepath); err != nil {
		return nil, err
	} else if fileInfo.IsDir() {
		return nil, errors.New("the file provided is a directory")
	}

	return os.Open(filepath)
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
	return nil
}
