package service

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestCronExpressions(t *testing.T) {
	c, err := cron.ParseStandard("0 0 1 1 6")
	assert.NoError(t, err, "Got error.")
	c, err = cron.ParseStandard("15-59/10,*/2 * * * *")
	assert.NoError(t, err, "Got error.")
	log.Debug().Any("c", c.Next(time.Now())).Msg("Next")
	c, err = cron.ParseStandard("15-59/10,*/2 * *    * *")
	assert.NoError(t, err, "Got error.")
}
