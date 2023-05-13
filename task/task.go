package task

import (
	"time"
)

type Task struct {
	Code      string
	Desc      string
	Notes     string
	StartDate time.Time
	EndDate   time.Time
}



