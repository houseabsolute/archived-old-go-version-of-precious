package pathfilter

import (
	"fmt"
	"path/filepath"
	"strings"

	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	gitignore "github.com/sabhiram/go-gitignore"
)

type Filter struct {
	include []string
	exclude []string
	// This is a map of directories to gitignore(-style) files. The directory
	// root will be stripped from a file when checking it with each ignorer so
	// that "/foo" style ignores are handled correctly.
	ignore map[string]*gitignore.GitIgnore
}

func New(include, exclude, ignoreFiles []string) (*Filter, error) {
	ignore := map[string]*gitignore.GitIgnore{}
	if ignoreFiles != nil {
		for _, f := range ignoreFiles {
			i, err := gitignore.CompileIgnoreFile(f)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("Could not compile gitignore style file at %s", f))
			}
			ignore[filepath.Dir(f)] = i
		}
	}

	return &Filter{include, exclude, ignore}, nil
}

func (f *Filter) ApplyAllRules(paths []string) ([]string, error) {
	filtered := []string{}
	for _, path := range paths {
		exclude, err := f.pathIsExcluded(path)
		if err != nil {
			return []string{}, err
		}
		if exclude {
			continue
		}

		include, err := f.pathIsIncluded(path)
		if err != nil {
			return []string{}, err
		}
		if !include {
			continue
		}

		filtered = append(filtered, path)
	}

	return filtered, nil
}

func (f *Filter) ApplyExcludeRules(paths []string) ([]string, error) {
	filtered := []string{}
	for _, path := range paths {
		exclude, err := f.pathIsExcluded(path)
		if err != nil {
			return []string{}, err
		}
		if exclude {
			continue
		}
		filtered = append(filtered, path)
	}

	return filtered, nil
}

func (f *Filter) pathIsExcluded(path string) (bool, error) {
	for d, i := range f.ignore {
		if strings.HasPrefix(d, path) {
			if i.MatchesPath(strings.TrimPrefix(path, d)) {
				return true, nil
			}
		}
	}

	for _, e := range f.exclude {
		matched, err := checkZglob(e, path)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

func (f *Filter) pathIsIncluded(path string) (bool, error) {
	for _, i := range f.include {
		matched, err := checkZglob(i, path)
		if err != nil {
			return false, err
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}

func checkZglob(pattern, path string) (bool, error) {
	matched, err := zglob.Match(pattern, path)
	if err != nil {
		return false, errors.Wrap(err,
			fmt.Sprintf("Error matching %s against exclude zglob pattern %s", path, pattern))
	}
	return matched, nil
}
