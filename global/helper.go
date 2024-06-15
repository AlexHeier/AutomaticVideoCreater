package global

import (
	"math/rand"
	"time"
)

func Random(array []string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return array[r.Intn(len(array))]
}
