package basepaths

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	alog "github.com/apex/log"
	"github.com/houseabsolute/precious/internal/pathfilter"
	"github.com/mattn/go-zglob/fastwalk"
	"github.com/pkg/errors"
)

type Mode int

const (
	FromCLI Mode = iota
	AllFiles
	GitModified
	GitStaged
)

type BasePaths struct {
	l         *alog.Logger
	mode      Mode
	cliPaths  []string
	basePaths *[]string
	filter    *pathfilter.Filter
}

func New(l *alog.Logger, m Mode, cliPaths, exclude, ignoreFiles []string) (*BasePaths, error) {
	if m != FromCLI && len(cliPaths) != 0 {
		return nil, errors.New("You cannot provide paths on the command line along with the -a, -g, or -s flags")
	}

	filter, err := pathfilter.New([]string{}, exclude, ignoreFiles)
	if err != nil {
		return nil, err
	}

	return &BasePaths{
		l:        l,
		mode:     m,
		cliPaths: cliPaths,
		filter:   filter,
	}, nil
}

func (bf *BasePaths) Paths() ([]string, error) {
	if bf.basePaths != nil {
		return *bf.basePaths, nil
	}

	start, err := bf.startingPaths()
	if err != nil {
		return []string{}, err
	}

	paths := []string{}
	for _, p := range start {
		fi, err := os.Stat(p)
		if err != nil {
			return []string{}, errors.Wrap(err, fmt.Sprintf("Could not stat path %s", p))
		}

		if fi.IsDir() {
			found, err := bf.searchDir(p)
			if err != nil {
				return []string{}, err
			}

			paths = append(paths, found...)
			continue
		}

		paths = append(paths, p)
	}

	paths, err = bf.filter.ApplyExcludeRules(paths)
	if err != nil {
		return []string{}, err
	}

	sort.Strings(paths)
	bf.basePaths = &paths

	return *bf.basePaths, nil
}

func (bf *BasePaths) startingPaths() ([]string, error) {
	if len(bf.cliPaths) > 0 {
		bf.l.Debugf("Using explicit list of starting paths: %s", bf.cliPaths)
		return bf.cliPaths, nil
	} else if bf.mode == GitModified {
		bf.l.Info("Using git modified paths as starting paths")
		return nil, nil
	} else if bf.mode == GitStaged {
		bf.l.Info("Using git staged paths as starting paths")
		return nil, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return []string{}, errors.Wrap(err, "Could not get your current working directory")
	}
	bf.l.Infof("Using %s as starting path", wd)
	return []string{wd}, nil
}

func (bf *BasePaths) searchDir(dir string) ([]string, error) {
	paths := []string{}
	err := fastwalk.FastWalk(dir, func(path string, typ os.FileMode) error {
		filtered, err := bf.filter.ApplyExcludeRules([]string{path})
		if err != nil {
			return err
		}
		paths = append(paths, filtered...)

		if len(paths) == 0 && typ.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})

	return paths, err
}

func (bf *BasePaths) UnstashIfNeeded() error {
	return nil
}
