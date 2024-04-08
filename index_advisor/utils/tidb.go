package utils

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type LocalTiDBServer struct {
	cmd    *exec.Cmd
	port   int
	tmpDir string
}

func (s *LocalTiDBServer) Release() error {
	Infof("Kill TiDB process pid: %v", s.cmd.Process.Pid)
	err := syscall.Kill(s.cmd.Process.Pid, syscall.SIGQUIT)
	if err != nil {
		return err
	}

	Infof("wait for TiDB to close")
	time.Sleep(time.Second * 3)

	Infof("Clean tmpDir: %v", s.tmpDir)
	os.RemoveAll(s.tmpDir)
	return nil
}

func (s *LocalTiDBServer) DSN() string {
	return fmt.Sprintf("root:@tcp(127.0.0.1:%v)/", s.port)
}

// StartLocalTiDBServer starts a TiDB server with the given version.
func StartLocalTiDBServer(ver string) (*LocalTiDBServer, error) {
	if ver == "" {
		ver = "nightly"
	}
	tiupPath, err := exec.LookPath("tiup")
	if err != nil {
		return nil, fmt.Errorf("failed to find tiup cmd: %v, please install tiup first: https://docs.pingcap.com/tidb/dev/tiup-overview", err)
	}

	port, err := GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get a free port: %v", err)
	}
	statusPort, err := GetFreePort()
	if err != nil {
		return nil, fmt.Errorf("failed to get a free port: %v", err)
	}
	tmpDir, err := GetTempDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get a temp dir: %v", err)
	}
	logFilePath := path.Join(tmpDir, "tidb.log")
	slowLogFilePath := path.Join(tmpDir, "tidb_slow.log")

	cmd := exec.Command(tiupPath, fmt.Sprintf("tidb:%v", ver),
		fmt.Sprintf("--status=%v", statusPort),
		fmt.Sprintf("-P=%v", port),
		fmt.Sprintf("--path=%v", tmpDir),
		fmt.Sprintf("--log-file=%v", logFilePath),
		fmt.Sprintf("--log-slow-query=%v", slowLogFilePath))
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	Infof("Starting TiDB %v", cmd.String())
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start TiDB: %v", err)
	}

	Infof("Wait for TiDB to start, pid: %v", cmd.Process.Pid)
	ok := false
	dsn := fmt.Sprintf("root:@tcp(127.0.0.1:%v)/test", port)
	for i := 0; i < 10; i++ {
		if PingLocalTiDB(dsn) {
			Infof("TiDB started, port: %v, tmpDir: %v", port, tmpDir)
			ok = true
			break
		}
		time.Sleep(time.Second * 2)
		Infof("Wait for TiDB to start, pid: %v", cmd.Process.Pid)
	}
	if !ok {
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to start TiDB, stderr: %v", stdErr.String())
	}

	return &LocalTiDBServer{
		cmd:    cmd,
		port:   port,
		tmpDir: tmpDir,
	}, nil
}

// GetTempDir returns an temporary directory path
func GetTempDir() (string, error) {
	return os.MkdirTemp("", "index_advisor_tidb_tmp")
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func PingLocalTiDB(dsn string) bool {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false
	}
	defer db.Close()
	return db.Ping() == nil
}
