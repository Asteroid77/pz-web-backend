package runtime

import (
	"os"
	"time"
)

type OSRuntime struct{}

func (OSRuntime) Executable() (string, error) { return os.Executable() }
func (OSRuntime) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
func (OSRuntime) Chmod(name string, mode os.FileMode) error { return os.Chmod(name, mode) }
func (OSRuntime) Create(name string) (*os.File, error)      { return os.Create(name) }

func (OSRuntime) Sleep(d time.Duration) { time.Sleep(d) }
func (OSRuntime) Exit(code int)         { os.Exit(code) }
