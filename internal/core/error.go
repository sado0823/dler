package core

import (
	"fmt"
	"github.com/pkg/errors"
	"os"
)

func ExitWithErr(err error) {
	if err != nil {
		fmt.Printf("%v\n", errors.Cause(err))
		os.Exit(1)
	}
}
