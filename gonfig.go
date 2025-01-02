package gonfig

import (
	"bytes"
	"os"
	"syscall"

	"gopkg.in/yaml.v3"
)

type ErrorCode string

const (
	NOT_ROOT           ErrorCode = "NOT_ROOT"
	READ_FAILED        ErrorCode = "READ_FAILED"
	PARSE_FAILED       ErrorCode = "PARSE_FAILED"
	SETTING_GID_FAILED ErrorCode = "SETTING_GID_FAILED"
	SETTING_UID_FAILED ErrorCode = "SETTING_UID_FAILED"
)

type Error struct {
	Code  ErrorCode
	Cause error
}

func (e *Error) Error() string {
	return string(e.Code)
}

type RunAsConfig interface {
	Uid() int
	Gid() int
}

func Get(path string, into RunAsConfig) error {
	if err := ensureRoot(); err != nil {
		return err
	}

	data, err := read(path)
	if err != nil {
		return err
	}

	if err := parse(data, into); err != nil {
		return err
	}

	if err := dropRoot(into.Uid(), into.Gid()); err != nil {
		return err
	}

	return nil
}

func ensureRoot() error {
	if os.Getuid() != 0 {
		return &Error{
			Code: NOT_ROOT,
		}
	} else {
		return nil
	}
}

func read(filename string) ([]byte, error) {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return nil, &Error{
			Code:  READ_FAILED,
			Cause: err,
		}
	}
	return contents, nil
}

func parse(data []byte, into RunAsConfig) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	if err := decoder.Decode(into); err != nil {
		return &Error{
			Code:  PARSE_FAILED,
			Cause: err,
		}
	}
	return nil
}

func dropRoot(uid int, gid int) error {
	if err := syscall.Setgid(gid); err != nil {
		return &Error{
			Code:  SETTING_GID_FAILED,
			Cause: err,
		}
	}
	if err := syscall.Setuid(uid); err != nil {
		return &Error{
			Code:  SETTING_UID_FAILED,
			Cause: err,
		}
	}
	return nil
}
