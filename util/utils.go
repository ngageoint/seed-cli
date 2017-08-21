package util

import (
	"fmt"
	"os"
	"time"
)

//TimeTrack function for timing function calls. Usage:
// defer TimeTrack(time.Now()) at the beginning of the timed function
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "%s took %s\n", name, elapsed)
}

//Exit type to handle exiting
type Exit struct{ Code int }

//HandleExit Looks at the panic for Exit codes vs actual panics
func HandleExit() {
	if e := recover(); e != nil {
		if exit, ok := e.(Exit); ok == true {
			os.Exit(exit.Code)
		}
		panic(e)
	}
}
