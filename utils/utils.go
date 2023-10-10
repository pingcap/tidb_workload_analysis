package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

// SaveContentTo saves the given content to the given file.
func SaveContentTo(fpath, content string) error {
	return os.WriteFile(fpath, []byte(content), 0644)
}

// FileExists tests whether this file exists and is or not a directory.
func FileExists(filename string) (exist, isDir bool) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, false
	}
	return true, info.IsDir()
}

func PrepareDir(dirPath string) error {
	if err := os.RemoveAll(dirPath); err != nil {
		return err
	}
	return os.MkdirAll(dirPath, 0755)
}

func ReadURL(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("get %v error: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %v error: status code is %v not OK(200)", url, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %v response body error: %v", url, err)
	}
	return data, nil
}

// GetDBNameFromDSN extracts the database name from the given DSN.
func GetDBNameFromDSN(dsn string) (dsnWithoutDB, dbName string) {
	idx := strings.Index(dsn, "/")
	if idx == -1 {
		return dsn, ""
	}
	return dsn[:idx+1], strings.TrimSpace(dsn[idx+1:])
}

// ParseStmtsFromDir parses raw Queries from the given directory.
// Each *.sql in this directory is parsed as a single Query.
func ParseStmtsFromDir(dirPath string) (sqls, fileNames []string, err error) {
	des, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range des {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		fpath := path.Join(dirPath, entry.Name())
		content, err := os.ReadFile(fpath)
		if err != nil {
			return nil, nil, err
		}
		sql := strings.TrimSpace(string(content))
		sqls = append(sqls, sql)
		fileNames = append(fileNames, entry.Name())
	}
	return
}

// ParseStmtsFromFile parses raw Queries from the given file.
// It ignore all comments, and assume all Queries are separated by ';'.
func ParseStmtsFromFile(fpath string) ([]string, error) {
	data, err := os.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	var filteredLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "--") { // empty line or comment
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	content := strings.Join(filteredLines, "\n")

	tmp := strings.Split(content, ";")
	var sqls []string
	for _, sql := range tmp {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}
		sqls = append(sqls, sql)
	}
	return sqls, nil
}

func Min[T int | float64](xs ...T) T {
	res := xs[0]
	for _, x := range xs {
		if x < res {
			res = x
		}
	}
	return res
}

func Max[T int | float64](xs ...T) T {
	res := xs[0]
	for _, x := range xs {
		if x > res {
			res = x
		}
	}
	return res
}
