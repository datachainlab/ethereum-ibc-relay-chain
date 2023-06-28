package ethereum

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	sendCheckpointFileName = "send.cp"
	recvCheckpointFileName = "recv.cp"
)

func (c *Chain) ensureDataDirectory() (string, error) {
	path := filepath.Join(c.homePath, "ethereum")
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return "", err
	}
	return path, nil
}

func (c *Chain) loadCheckpoint(fileName string) (uint64, error) {
	dir, err := c.ensureDataDirectory()
	if err != nil {
		return 0, err
	}

	bz, err := ioutil.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		return 0, err
	} else if len(bz) != 8 {
		return 0, fmt.Errorf("Unexpected send checkpoint file size: %v", len(bz))
	}

	return binary.BigEndian.Uint64(bz), nil
}

func (c *Chain) saveCheckpoint(v uint64, fileName string) error {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, v)

	dir, err := c.ensureDataDirectory()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(dir, fileName), bz, os.ModePerm)
}
