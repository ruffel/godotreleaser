package terminal

import (
	"fmt"
	"os"
)

func Send(x fmt.Stringer) {
	fmt.Fprintln(os.Stdout, x)
}
