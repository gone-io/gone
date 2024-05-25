package priest

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/tools/go/packages"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	PriestFn FnKind = iota
	NewGonerFn
)

type Fn struct {
	Name string
	Kind FnKind
}

func (f Fn) Gen(pkgName string) string {
	switch f.Kind {
	case PriestFn:
		return fmt.Sprintf("%s%s(cemetery)", pkgName, f.Name)
	//case NewGonerFn:
	default:
		return fmt.Sprintf("cemetery.Bury(%s%s())", pkgName, f.Name)
	}
}

type parseResult struct {
	Path        string
	PkgName     string
	InjectNames []Fn
}

const InjectTag = "gone"

var packageReg = regexp.MustCompile("^package ([a-zA-Z][a-zA-Z0-9_]*)")
var injectReg = regexp.MustCompile(fmt.Sprintf("^//go:%s(\\s+.*|$)", InjectTag))
var funcReg = regexp.MustCompile("^func\\s+([A-Z][a-zA-Z0-9_]*)\\s*\\((.*?)\\)")
var priestParamReg = regexp.MustCompile("^([a-zA-Z0-9_]*)\\s*gone\\.Cemetery$")

type FnKind int

func goFileParse(goFilepath string) (*parseResult, error) {
	file, err := os.Open(goFilepath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	var PkgName string

	var InjectFns = make([]Fn, 0)

	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if PkgName == "" {
			match := packageReg.FindSubmatch(line)
			if len(match) == 2 {
				PkgName = string(match[1])
			}
		} else {
			if injectReg.Match(line) {
				line, _, err := reader.ReadLine()
				if err == io.EOF {
					return nil, errors.New("file unexpected end")
				}
				match := funcReg.FindSubmatch(line)
				if len(match) == 3 {
					paramStr := strings.TrimSpace(string(match[2]))

					if paramStr == "" {
						InjectFns = append(InjectFns, Fn{
							Name: string(match[1]),
							Kind: NewGonerFn,
						})
					}

					if priestParamReg.MatchString(paramStr) {
						InjectFns = append(InjectFns, Fn{
							Name: string(match[1]),
							Kind: PriestFn,
						})
					}
				}
			}
		}
	}

	if PkgName == "" || len(InjectFns) == 0 {
		return nil, nil
	}
	absPath, _ := filepath.Abs(goFilepath)

	return &parseResult{
		Path:        filepath.Dir(absPath),
		PkgName:     PkgName,
		InjectNames: InjectFns,
	}, nil
}

func goModuleInfo(dir string) (moduleName string, moduleAbsPath string, err error) {
	defer TimeStat("goModuleInfo:" + dir)()

	cfg := &packages.Config{
		Mode:  packages.NeedModule,
		Dir:   dir,
		Tests: false,
	}
	packageList, err := packages.Load(cfg, "")

	if err != nil {
		return "", "", err
	}

	p := packageList[0]

	if p.Module == nil {
		file, _ := os.ReadDir(dir)
		if len(file) == 0 {
			err = errors.New("do not found .go file")
			return
		}

		for _, f := range file {
			if f.IsDir() && f.Name() != ".git" {
				moduleName, moduleAbsPath, err = goModuleInfo(path.Join(dir, f.Name()))
				if err == nil {
					return
				}
			}
		}
		err = errors.New("do not found .go file")
		return
	}

	moduleName, moduleAbsPath = p.Module.Path, p.Module.Dir
	return
}
