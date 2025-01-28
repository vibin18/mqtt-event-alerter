package core

import (
	"time"
)

type Alert struct {
	DateTime time.Time
	Object   string
}
