package ethereum

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

	bz, err := ioutil.ReadFile(filepath.Join(dir, checkpointFileName(cpType)))
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
	} else if len(bz) != 8 {
		return 0, fmt.Errorf("Unexpected send checkpoint file size: %v", len(bz))
	}

	return binary.BigEndian.Uint64(bz), nil
}

func (c *Chain) saveCheckpoint(v uint64, cpType checkpointType) error {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, v)

	dir, err := c.ensureDataDirectory()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(dir, checkpointFileName(cpType)), bz, os.ModePerm)
}
