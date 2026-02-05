package updateapp

import (
	"io"
	"net/http"
	"time"

	"pz-web-backend/internal/infra/runtime"
	sysupdate "pz-web-backend/internal/system/update"
)

type Service struct {
	DevMode bool

	Checker sysupdate.Service

	HTTPClient *http.Client
	TmpPath    string
	Runtime    runtime.Ops
}

func NewService(devMode bool, checker sysupdate.Service) Service {
	return Service{
		DevMode:    devMode,
		Checker:    checker,
		HTTPClient: http.DefaultClient,
		TmpPath:    "/tmp/pz-config-app.new",
		Runtime:    runtime.OSRuntime{},
	}
}

func (s Service) CheckUpdate() (string, string, error) {
	checker := s.Checker
	if checker.HTTPClient == nil {
		checker.HTTPClient = s.httpClient()
	}
	return checker.CheckUpdate()
}

func (s Service) PerformUpdate(downloadURL string) error {
	client := s.httpClient()
	resp, err := client.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := s.Runtime.Create(s.TmpPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	if err := s.Runtime.Chmod(s.TmpPath, 0o755); err != nil {
		return err
	}

	if s.DevMode {
		// 开发模式不覆盖当前运行的二进制。
		return nil
	}

	binPath, err := s.Runtime.Executable()
	if err != nil {
		binPath = "/opt/pz-web-backend/pz-web-backend"
	}

	_ = s.Runtime.Rename(binPath, binPath+".bak")
	if err := s.Runtime.Rename(s.TmpPath, binPath); err != nil {
		_ = s.Runtime.Rename(binPath+".bak", binPath)
		return err
	}

	go func() {
		s.Runtime.Sleep(1 * time.Second)
		s.Runtime.Exit(0)
	}()

	return nil
}

func (s Service) httpClient() *http.Client {
	if s.HTTPClient != nil {
		return s.HTTPClient
	}
	return http.DefaultClient
}
