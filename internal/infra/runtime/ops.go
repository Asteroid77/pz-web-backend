package runtime

import (
	"os"
	"time"
)

type Ops interface {
	Executable() (string, error)
	Rename(oldpath, newpath string) error
	Chmod(name string, mode os.FileMode) error
	Create(name string) (*os.File, error)

	Sleep(d time.Duration)
	Exit(code int)
}
