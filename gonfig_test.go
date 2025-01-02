package gonfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testConfig struct {
	Db struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"db"`
	RunAs struct {
		Uid int `yaml:"uid"`
		Gid int `yaml:"gid"`
	} `yaml:"runas"`
}

func (tc *testConfig) Gid() int {
	return tc.Gid()
}

func (tc *testConfig) Uid() int {
	return tc.Uid()
}

func TestParse(t *testing.T) {
	data, err := read("test_config.yaml")
	if err != nil {
		t.Error(err)
	}
	tc := testConfig{}
	if err := parse(data, &tc); err != nil {
		t.Error(err)
	}

	assert.Equal(t, "dbuser", tc.Db.User)
	assert.Equal(t, "dbpasswd", tc.Db.Password)
	assert.Equal(t, 1000, tc.RunAs.Uid)
	assert.Equal(t, 1000, tc.RunAs.Gid)
}
