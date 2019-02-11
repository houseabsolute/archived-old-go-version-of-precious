package tidymaster

import (
	"github.com/apex/log"
	"github.com/houseabsolute/precious/internal/basepaths"
	"github.com/houseabsolute/precious/internal/config"
)

type TidyMaster struct {
	l  *log.Logger
	c  *config.Config
	bf *basepaths.BasePaths
}

func New(l *log.Logger, c *config.Config, bf *basepaths.BasePaths) (*TidyMaster, error) {
	return &TidyMaster{l, c, bf}, nil
}

func (tm *TidyMaster) Tidy() error {
	// for _, filter := range tm.c.Tidiers() {
	// 	filter.Tidy(bf)
	// }

	return nil
}
