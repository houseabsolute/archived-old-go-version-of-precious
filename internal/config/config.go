package config

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	alog "github.com/apex/log"
	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
)

type server struct {
	port int64
}

type command struct {
	pathFlag    string
	okExitCodes []int64
}

type filterType string

const (
	tidy filterType = "tidy"
	lint            = "lint"
	both            = "both"
)

type filterConfig struct {
	name    string
	ignore  []string
	include []string
	exclude []string
	typ     string
	cmd     []string
	args    []string
	onDir   bool
	server  *server
	command *command
}

type Config struct {
	Ignore  []string
	Exclude []string
	filters []filterConfig
	l       *alog.Logger
}

func NewFromFile(l *alog.Logger, file string) (*Config, error) {
	tree, err := toml.LoadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("Error reading config from %s", file))
	}

	c := &Config{l: l}
	msgs := validateAndSetConfig(l, c, tree, file)
	if len(msgs) != 0 {
		combined := fmt.Sprintf("There was one or more errors with your configuration file at %s:\n", file)
		for _, M := range msgs {
			combined += M + "\n"
		}
		return nil, errors.New(combined)
	}

	return c, nil
}

func validateAndSetConfig(l *alog.Logger, c *Config, tree *toml.Tree, file string) []string {
	msgs := []string{}

	c.Ignore = getStringOrStringArray("global", tree, "ignore", &msgs)
	c.Exclude = getStringOrStringArray("global", tree, "exclude", &msgs)
	c.filters = getFilters(l, tree, file, &msgs)

	return msgs
}

func getFilters(l *alog.Logger, tree *toml.Tree, file string, msgs *[]string) []filterConfig {
	if !tree.Has("servers") && !tree.Has("commands") {
		*msgs = append(*msgs, fmt.Sprintf("You must define at least one server or command in your config file at %s", file))
		return []filterConfig{}
	}

	filters := map[int]filterConfig{}

	configRoot, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		*msgs = append(*msgs, fmt.Sprintf("Error getting abs path for %s", file))
		return []filterConfig{}
	}

	if tree.Has("servers") {
		l.Debug("Found [[servers]] in config")

		servers := tree.Get("servers")
		switch s := servers.(type) {
		case []*toml.Tree:
			for _, t := range s {
				line := t.Position().Line
				name := t.Keys()[0]
				l.Debugf("Found server %s at line %d", name, line)
				filters[line] = treeToServer(l, configRoot, name, t.Get(name).([]*toml.Tree)[0], msgs)
			}
		default:
			*msgs = append(*msgs,
				fmt.Sprintf("The servers in the config file at %s must be an array of tables ([[servers]])", file))
		}
	}

	if tree.Has("commands") {
		l.Debug("Found [[commands]] in config")

		commands := tree.Get("commands")
		switch c := commands.(type) {
		case []*toml.Tree:
			for _, t := range c {
				line := t.Position().Line
				name := t.Keys()[0]
				l.Debugf("Found command %s at line %d", name, line)
				filters[t.Position().Line] = treeToCommand(l, configRoot, name, t.Get(name).([]*toml.Tree)[0], msgs)
			}
		default:
			*msgs = append(*msgs,
				fmt.Sprintf("The commands in the config file at %s must be an array of tables ([[commands]])", file))
		}
	}

	keys := []int{}
	for p := range filters {
		keys = append(keys, p)
	}
	sort.Ints(keys)

	sorted := []filterConfig{}
	for _, k := range keys {
		sorted = append(sorted, filters[k])
	}

	return sorted
}

func treeToServer(l *alog.Logger, configRoot, name string, s *toml.Tree, msgs *[]string) filterConfig {
	f := baseFilterConfig(configRoot, name, s, msgs)
	f.server = &server{port: getInt64(name, s, "port", msgs)}
	l.Debugf("%+v", f)
	return f
}

func treeToCommand(l *alog.Logger, configRoot, name string, c *toml.Tree, msgs *[]string) filterConfig {
	f := baseFilterConfig(configRoot, name, c, msgs)
	f.command = &command{
		pathFlag:    getString(name, c, "path_flag", msgs),
		okExitCodes: getInt64OrInt64Array(name, c, "ok_exit_codes", msgs),
	}
	l.Debugf("%+v", f)
	return f
}

func baseFilterConfig(configRoot, name string, t *toml.Tree, msgs *[]string) filterConfig {
	return filterConfig{
		name:    name,
		ignore:  getStringOrStringArray(name, t, "ignore", msgs),
		exclude: getStringOrStringArray(name, t, "exclude", msgs),
		include: getStringOrStringArray(name, t, "include", msgs),
		typ:     getString(name, t, "type", msgs),
		cmd:     getStringOrStringArray(name, t, "cmd", msgs),
		args:    applyRoot(configRoot, getStringOrStringArray(name, t, "args", msgs)),
		onDir:   getBool(name, t, "on_dir", msgs),
	}
}

func getString(name string, tree *toml.Tree, key string, msgs *[]string) string {
	if !tree.Has(key) {
		return ""
	}

	raw := tree.Get(key)
	if val, ok := raw.(string); ok {
		return val
	}

	*msgs = append(*msgs, fmt.Sprintf("The %s.%s key must be a string, not a %s", name, key, reflect.TypeOf(raw)))

	return ""
}

func getStringOrStringArray(name string, tree *toml.Tree, key string, msgs *[]string) []string {
	if !tree.Has(key) {
		return []string{}
	}

	raw := tree.Get(key)
	switch val := raw.(type) {
	case string:
		return []string{val}
	case []interface{}:
		vals := []string{}
		for _, r := range val {
			if v, ok := r.(string); ok {
				vals = append(vals, v)
			} else {
				vals = []string{}
				break
			}
		}
		if len(vals) != 0 {
			return vals
		}
	}

	*msgs = append(*msgs, fmt.Sprintf("The %s %s key must be a string or array of strings, not a %s", name, key, reflect.TypeOf(raw)))

	return []string{}
}

func getBool(name string, tree *toml.Tree, key string, msgs *[]string) bool {
	if !tree.Has(key) {
		return false
	}

	raw := tree.Get(key)
	if val, ok := raw.(bool); ok {
		return val
	}

	*msgs = append(*msgs, fmt.Sprintf("The %s.%s key must be a bool, not a %s", name, key, reflect.TypeOf(raw)))

	return false
}

func getInt64(name string, tree *toml.Tree, key string, msgs *[]string) int64 {
	if !tree.Has(key) {
		return 0
	}

	raw := tree.Get(key)
	if val, ok := raw.(int64); ok {
		return val
	}

	*msgs = append(*msgs, fmt.Sprintf("The %s.%s key must be an int, not a %s", name, key, reflect.TypeOf(raw)))

	return 0
}

func getInt64OrInt64Array(name string, tree *toml.Tree, key string, msgs *[]string) []int64 {
	if !tree.Has(key) {
		return []int64{}
	}

	raw := tree.Get(key)
	switch val := raw.(type) {
	case int64:
		return []int64{val}
	case []interface{}:
		vals := []int64{}
		for _, r := range val {
			if v, ok := r.(int64); ok {
				vals = append(vals, v)
			} else {
				vals = []int64{}
				break
			}
		}
		if len(vals) != 0 {
			return vals
		}
	}

	*msgs = append(*msgs, fmt.Sprintf("The %s %s key must be an int or array of ints, not a %s", name, key, reflect.TypeOf(raw)))

	return []int64{}
}

func applyRoot(root string, vals []string) []string {
	applied := []string{}
	for _, v := range vals {
		applied = append(applied, strings.Replace(v, "$CONFIG_DIR", root, -1))
	}
	return applied
}

func (c *Config) Tidyers() []tidyer.Tidyer {
	tidyers := []*tidyer.Tidyer{}
	for _, f := range c.filters {
		if f.typ != lint {
			var t *tidyer.Tidyer
			var err error
			if f.server != nil {
				t, err = tidyer.NewServer(
					f.name,
					f.ignore,
					f.include,
					f.exclude,
					f.cmd,
					f.args,
					f.onDir,
					f.server.port,
				)
			} else {
				t, err = tidyer.NewCommand(
					f.name,
					f.ignore,
					f.include,
					f.exclude,
					f.cmd,
					f.args,
					f.onDir,
					f.command.pathFlag,
					f.command.okExitCodes,
				)
			}

			tidyers = append(tidyers, t)
		}
	}
	return t
}
