package testing

import (
	"testing"

	"github.com/pbloigu/gonfig/api"
	"github.com/pbloigu/gonfig/server/configurations"
	"github.com/pbloigu/gonfig/server/service"
	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	cMock := NewMockConfigurations(t)
	cMock.On("ListCronTriggers").Return([]configurations.CronTrigger{})
	s := service.New(NewMockSeriesDb(t), cMock)

	r := api.CronValidationRequest{
		Minute: "0",
		Hour:   "0",
		Dom:    "1",
		Month:  "1",
		Dow:    "6",
	}

	assert.True(t, s.IsValid(r))

	r = api.CronValidationRequest{
		Minute: "15-59/10,*/2",
		Hour:   "*",
		Dom:    "*",
		Month:  "*",
		Dow:    "*",
	}
	assert.True(t, s.IsValid(r))

	r = api.CronValidationRequest{
		Minute: "15-59/10,*/2",
		Hour:   "*",
		Dom:    "*",
		Month:  "*",
		Dow:    "*",
	}
	assert.True(t, s.IsValid(r))

}

func TestNewPagination(t *testing.T) {
	cMock := NewMockConfigurations(t)
	cMock.On("ListCronTriggers").Return([]configurations.CronTrigger{})
	s := service.New(NewMockSeriesDb(t), cMock)

	p := s.NewPagination(10, 5, 3)
	assert.Equal(t, p.Page, 3)
	assert.Equal(t, p.Size, 10)

	p = s.NewPagination(0, 5, 0)
	assert.Equal(t, p.Page, 1)
	assert.Equal(t, p.Size, 5)

	p = s.NewPagination(-1, 5, 0)
	assert.Equal(t, p.Page, 1)
	assert.Equal(t, p.Size, 5)
}

func TestNewSort(t *testing.T) {
	cMock := NewMockConfigurations(t)
	cMock.On("ListCronTriggers").Return([]configurations.CronTrigger{})
	s := service.New(NewMockSeriesDb(t), cMock)

	srt := s.NewSort("foo", "default", "DESC")
	assert.Equal(t, "foo", srt.Sort)
	assert.Equal(t, service.DESC, srt.Dir)
}
