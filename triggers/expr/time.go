package expr

import (
	"time"
)

func init() {
	helpers := map[string]interface{}{
		"parse": toTime,
		"now":   now,
	}
	register("time", helpers)
}

func toTime(timestamp string) time.Time {
	res, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		panic(err)
	}
	return res
}

func now() time.Time {
	return time.Now()
}
