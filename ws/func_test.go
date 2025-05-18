package ws

import (
	"testing"

	"github.com/rs/zerolog/log"
)

func TestTest(t *testing.T) {
	m := make(map[int]func())

	for i := range 2 {
		m[i] = func() {
			log.Debug().Any("i", i).Msg("Called!")
		}
	}

	m[0]()
	m[1]()
}
