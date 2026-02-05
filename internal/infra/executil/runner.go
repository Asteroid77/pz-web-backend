package executil

type Runner interface {
	CombinedOutput(name string, args ...string) ([]byte, error)
}
