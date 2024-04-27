package pool_test

import (
	"fmt"
	"testing"

	"github.com/ysoding/spidey/pool"
)

type theWork struct {
}

func (*theWork) Work(context interface{}) {
	fmt.Printf("%s : Performing Work\n", context)
}

func TestPool(t *testing.T) {
	pool := pool.New()

	pool.Do("TEST", &theWork{})
	pool.Do("TEST", &theWork{})
	pool.Shutdown()
}
