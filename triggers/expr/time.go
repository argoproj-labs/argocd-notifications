package expr

import (
	"time"
)

func init() {
	register("time", newTimeExprs())
}

func newTimeExprs() map[string]interface{} {
	return map[string]interface{}{
		"Parse": parse,
		"Now":   now,
	}
}

func parse(timestamp string) time.Time {
	res, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		panic(err)
	}
	return res
}

func now() time.Time {
	return time.Now()
}
