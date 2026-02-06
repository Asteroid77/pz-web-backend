package supervisor

type Restarter interface {
	RestartPZServer() error
}
