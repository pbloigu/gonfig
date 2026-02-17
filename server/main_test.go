package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/mariadb"
)

func TestMain(t *testing.T) {
	tCtx := t.Context()
	tc, err := mariadb.Run(tCtx,
		"mariadb:11.0.3",
	)
	if err != nil {
		t.Error(err)
	}
	cstr, err := tc.ConnectionString(tCtx)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%s", cstr)
	os.Args = []string{"whatever", "--seriesDb", cstr, "--frontendPort", "6677", "--logLevel", "debug"}
	main()
}
