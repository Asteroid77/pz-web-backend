package supervisor

import "pz-web-backend/internal/infra/executil"

type SupervisorctlRestarter struct {
	Runner executil.Runner
}

func (r SupervisorctlRestarter) RestartPZServer() error {
	_, err := r.Runner.CombinedOutput(
		"/usr/bin/supervisorctl",
		"-c",
		"/etc/supervisor/conf.d/supervisord.conf",
		"restart",
		"pzserver",
	)
	return err
}
