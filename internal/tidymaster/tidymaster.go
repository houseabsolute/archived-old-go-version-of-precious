package tidymaster

import (
	alog "github.com/apex/log"
	"github.com/houseabsolute/precious/internal/basepaths"
	"github.com/houseabsolute/precious/internal/config"
)

type TidyMaster struct {
	l  *alog.Logger
	c  *config.Config
	bp *basepaths.BasePaths
}

func New(l *alog.Logger, c *config.Config, bp *basepaths.BasePaths) (*TidyMaster, error) {
	return &TidyMaster{l, c, bp}, nil
}

func (tm *TidyMaster) Tidy() error {
	paths := tm.bp.Paths()

	for _, filter := range tm.c.Tidiers() {
		filter.Tidy(paths)
	}

	return nil
}
