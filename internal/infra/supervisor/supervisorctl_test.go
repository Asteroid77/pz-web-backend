package supervisor

import (
	"errors"
	"reflect"
	"testing"
)

type fakeRunner struct {
	name string
	args []string

	out []byte
	err error
}

func (r *fakeRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	r.name = name
	r.args = append([]string(nil), args...)
	return r.out, r.err
}

func TestSupervisorctlRestarter_RestartPZServer_CallsExpectedCommand(t *testing.T) {
	runner := &fakeRunner{}
	restarter := SupervisorctlRestarter{Runner: runner}

	if err := restarter.RestartPZServer(); err != nil {
		t.Fatalf("err=%v", err)
	}

	if runner.name != "/usr/bin/supervisorctl" {
		t.Fatalf("name=%q", runner.name)
	}
	wantArgs := []string{
		"-c",
		"/etc/supervisor/conf.d/supervisord.conf",
		"restart",
		"pzserver",
	}
	if !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("args=%v want %v", runner.args, wantArgs)
	}
}

func TestSupervisorctlRestarter_RestartPZServer_PropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	restarter := SupervisorctlRestarter{Runner: &fakeRunner{err: wantErr}}

	if err := restarter.RestartPZServer(); err == nil {
		t.Fatalf("expected error")
	}
}
