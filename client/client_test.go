package client

import (
	"log"
	"testing"
)

func TestCC(t *testing.T) {
	t.Setenv("GONFIG_HOST", "localhost")
	t.Setenv("GONFIG_APPID", "c56a0884-1c4c-4dd7-b846-dc848a49b38d")
	t.Setenv("GONFIG_APIKEY", "ae101fe9-16b6-4f08-93be-c0c22611e8f1")
	t.Setenv("REST_PORT", "8081")
	t.Setenv("CC_PORT", "9000")
	ctx := t.Context()

	t.Run("TestCC", func(t *testing.T) {
		_, err := New(ctx, log.Default(), nil)
		if err != nil {
			t.Errorf("%v", err)
		}
	})
}
