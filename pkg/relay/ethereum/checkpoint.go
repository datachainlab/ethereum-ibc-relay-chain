package ethereum

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type checkpointType string

const (
	sendCheckpoint checkpointType = "send"
	recvCheckpoint checkpointType = "recv"
)

func checkpointFileName(cpType checkpointType) string {
	return string(cpType) + ".cp"
}

func (c *Chain) ensureDataDirectory() (string, error) {
	path := filepath.Join(c.homePath, "ethereum")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}

func (c *Chain) loadCheckpoint(cpType checkpointType) (uint64, error) {
	dir, err := c.ensureDataDirectory()
	if err != nil {
		return 0, err
	}

	bz, err := os.ReadFile(filepath.Join(dir, checkpointFileName(cpType)))
	if err != nil {
		if os.IsNotExist(err) {
			switch cpType {
			case sendCheckpoint:
				return c.config.InitialSendCheckpoint, nil
			case recvCheckpoint:
				return c.config.InitialRecvCheckpoint, nil
			default:
				panic(fmt.Sprintf("unexpected checkpoint type: %v", cpType))
			}
		}
		return 0, err
	}

	return strconv.ParseUint(string(bz), 10, 64)
}

func (c *Chain) saveCheckpoint(v uint64, cpType checkpointType) error {
	bz := []byte(strconv.FormatUint(v, 10))

	dir, err := c.ensureDataDirectory()
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, checkpointFileName(cpType)), bz, os.ModePerm)
}
