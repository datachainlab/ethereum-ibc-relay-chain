package ethereum

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type ErrorRepository map[[4]byte]abi.Error

func defaultErrorABIs() ([]abi.Error, error) {
	strT, err := abi.NewType("string", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create string type: %v", err)
	}
	uintT, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uint256 type: %v", err)
	}

	return []abi.Error{
		abi.NewError("Error", []abi.Argument{{Name: "desc", Type: strT}}),
		abi.NewError("Panic", []abi.Argument{{Name: "reason", Type: uintT}}),
	}, nil
}

func CreateErrorRepository(abiPaths []string) (ErrorRepository, error) {
	var errABIs []abi.Error

	for _, dir := range abiPaths {
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error encountered during the file tree walk: err=%v, path=%s", err, path)
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
				f, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open a file: err=%v, path=%s", err, path)
				}
				defer f.Close()

				contractABI, err := abi.JSON(f)
				if err != nil {
					return fmt.Errorf("failed to parse a ABI JSON file: err=%v, path=%s", err, path)
				}

				for _, errABI := range contractABI.Errors {
					errABIs = append(errABIs, errABI)
				}
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("failed to read ABIs from a directory: err=%v, dirpath=%s", err, dir)
		}
	}

	return NewErrorRepository(errABIs)
}

func NewErrorRepository(errABIs []abi.Error) (ErrorRepository, error) {
	defaultErrABIs, err := defaultErrorABIs()
	if err != nil {
		return nil, fmt.Errorf("failed to create default error ABIs: %v", err)
	}

	errABIs = append(errABIs, defaultErrABIs...)

	repo := make(ErrorRepository)
	for _, errABI := range errABIs {
		if err := repo.Add(errABI); err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func (r ErrorRepository) Add(errABI abi.Error) error {
	var sel [4]byte
	copy(sel[:], errABI.ID[:4])
	if existingErrABI, ok := r[sel]; ok {
		if existingErrABI.Sig == errABI.Sig {
			return nil
		}
		return fmt.Errorf("error selector collision: selector=%x, newErrABI=%v, existingErrABI=%v", sel, errABI, existingErrABI)
	}
	r[sel] = errABI
	return nil
}

func (r ErrorRepository) Get(errorData []byte) (abi.Error, error) {
	if len(errorData) < 4 {
		return abi.Error{}, fmt.Errorf("the size of error data is less than 4 bytes: errorData=%x", errorData)
	}
	var sel [4]byte
	copy(sel[:], errorData[:4])
	errABI, ok := r[sel]
	if !ok {
		return abi.Error{}, fmt.Errorf("error ABI not found: errorData=%x", errorData)
	}
	return errABI, nil
}

func errorToJSON(errVal interface{}, errABI abi.Error) (string, error) {
	errVals, ok := errVal.([]interface{})
	if !ok {
		return "", fmt.Errorf("error value has unexpected type: expected=[]interface{}, actual=%T", errVal)
	}

	m := make(map[string]interface{})
	for i, v := range errVals {
		m[errABI.Inputs[i].Name] = v
	}

	bz, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal error value: %v", err)
	}

	return string(bz), nil
}

func (r ErrorRepository) ParseError(errorData []byte) (string, error) {
	errABI, err := r.Get(errorData)
	if err != nil {
		return "", fmt.Errorf("failed to find error ABI: %v", err)
	}
	errVal, err := errABI.Unpack(errorData)
	if err != nil {
		return "", fmt.Errorf("failed to unpack error: %v", err)
	}
	errStr, err := errorToJSON(errVal, errABI)
	if err != nil {
		return "", fmt.Errorf("failed to marshal error inputs into JSON: %v", err)
	}
	return fmt.Sprintf("%s%s", errABI.Name, errStr), nil
}
