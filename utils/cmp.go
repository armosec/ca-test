package utils

import (
	"time"

	"github.com/google/go-cmp/cmp"
)

var IgnoreTimeOption = cmp.FilterValues(func(x, y time.Time) bool { return true }, cmp.Ignore())
