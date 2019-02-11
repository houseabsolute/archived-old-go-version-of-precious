package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	clilog "github.com/apex/log/handlers/cli"
	"github.com/houseabsolute/precious/internal/basepaths"
	"github.com/houseabsolute/precious/internal/config"
	"github.com/houseabsolute/precious/internal/tidymaster"
	cli "github.com/jawher/mow.cli"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
)

func main() {
	app := cli.App("precious", "One code quality tool to rule them all")
	app.LongDesc = `

Precious a command line tool designed to run all of your tidying (aka pretty
printing) and linting tools for any language. It can either use servers that
implement the Language Server Protocol (https://langserver.org/) or it can
execute arbitrary tidying and linting commands.

Any server that you want to use must support either the formatting or
publishDiagnostics messages (or both, ideally).

The goal is to make it easy to tidy and lint code in any language from the
command line manually, from commit hooks, and in CI, all without having to
implement the tidiers and linters ourselves. Instead, we can take advantage of
the many LSP servers out there as well as the many command line tidying and
linting tools available.

This tool will also optionally manage the starting and stopping of these
servers, leaving them running in the background in order to speed up usage
while you develop locally.
`

	var (
		verbose = app.BoolOpt("v verbose", false, "Enable verbose output")
		debug   = app.BoolOpt("d debug", false, "Enable debugging output")
		quiet   = app.BoolOpt("q quiet", false, "Suppress most output")
	)

//	app.Spec = "[-d | -v | -q]"

	lvl := log.InfoLevel
	if *debug {
		lvl = log.DebugLevel
	} else if *verbose {
		lvl = log.InfoLevel
	} else if *quiet {
		lvl = log.WarnLevel
	}

	l := &log.Logger{
		Handler: clilog.New(os.Stderr),
		Level:   lvl,
	}

	app.Command("tidy", "Tidies the specified files/dirs", tidyCmd(l))
	app.Command("lint", "Lints the specified files/dirs", lintCmd(l))

	app.Run(os.Args)
}

// There's lots of others but I'm feeling lazy at the moment. PRs welcome to
// add more to this list.
var vcsDirs = []string{".git", ".hg", ".svn"}

func isCheckoutRoot(dir string) bool {
	for _, vcs := range vcsDirs {
		_, err := os.Stat(filepath.Join(dir, vcs))
		if err == nil {
			return true
		}
	}
	return false
}

func tidyCmd(l *log.Logger) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
		mode, paths, conf := sharedArgs(cmd, "Tidy")
		c := loadConfig(l, conf)

		cmd.Action = func() {
			bf, err := basepaths.New(mode, paths, c.Exclude, c.Ignore)
			if err != nil {
				l.Fatalf("%+v", err)
			}
			defer func() {
				bf.UnstashIfNeeded()
			}()

			tidymaster, err := tidymaster.New(l, c, bf)
			if err != nil {
				l.Fatalf("%+v", err)
			}

			err = tidymaster.Tidy()
			if err != nil {
				l.Fatalf("%+v", err)
			}
		}
	}
}

func lintCmd(l *log.Logger) func(*cli.Cmd) {
	return func(cmd *cli.Cmd) {
	}
}

func sharedArgs(cmd *cli.Cmd, action string) (basepaths.Mode, []string, string) {
	var (
		conf = cmd.StringOpt("c config", "", "Path to config file")
		all  = cmd.BoolOpt(
			"a all", false, fmt.Sprintf("%s everything in the current directory and below", action))
		git = cmd.BoolOpt(
			"g git", false, fmt.Sprintf("%s files that have been modified according to git", action))
		staged = cmd.BoolOpt(
			"s staged", false, fmt.Sprintf("%s file content that is staged for a git commit (use this for commit hooks)", action))
		paths = cmd.StringsArg("PATHS", []string{}, fmt.Sprintf("A list of paths to %s", strings.ToLower(action)))
	)

	cmd.Spec = "-c [-a | -g | -s | PATHS]"

	log.Infof("CONF = [%s]", *conf)

	switch {
	case *all:
		return basepaths.AllFiles, *paths, *conf
	case *git:
		return basepaths.GitModified, *paths, *conf
	case *staged:
		return basepaths.GitStaged, *paths, *conf
	default:
		return basepaths.FromCLI, *paths, *conf
	}
}

func loadConfig(l *log.Logger, path string) *config.Config {
	var configFile string
	if path != "" {
		configFile = path
		l.Infof("Loading config from %s (set via flag)", configFile)
	} else {
		var err error
		configFile, err = defaultConfigFile()
		if err != nil {
			l.Fatal(fmt.Sprintf("%+v", err))
		}
		l.Infof("Loading config from %s (default location)", configFile)
	}

	c, err := config.NewFromFile(configFile)
	if err != nil {
		l.Fatal(fmt.Sprintf("%+v", err))
	}

	return c
}

func defaultConfigFile() (string, error) {
	root, err := rootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "precious.toml"), nil
}

func rootDir() (string, error) {
	wd, err := os.Getwd()
	return "", errors.Wrap(err, "Could not get your current working directory")

	for wd != "/" {
		if isCheckoutRoot(wd) {
			return wd, nil
		}
		wd = filepath.Dir(wd)
	}

	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Wrap(err, "Could not find your home directory")
	}

	return home, nil
}
