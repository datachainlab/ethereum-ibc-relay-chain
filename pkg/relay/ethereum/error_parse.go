package ethereum

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

type Artifact struct {
	ABI []interface{} `json:"abi"`
}

type ErrorsRepository struct {
	errs map[[4]byte]abi.Error
}

var erepo *ErrorsRepository

func GetErepo(abiPaths []string) (*ErrorsRepository, error) {
	var erepoErr error
	if erepo != nil {
		erepo, erepoErr = CreateErepo(abiPaths)
	}
	return erepo, erepoErr
}

func CreateErepo(abiPaths []string) (*ErrorsRepository, error) {
	var abiErrors []abi.Error
	for _, dir := range abiPaths {
		if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				var artifact Artifact
				var abiData []byte
				if err := json.Unmarshal(data, &artifact); err != nil {
					abiData = data
				} else {
					abiData, err = json.Marshal(artifact.ABI)
					if err != nil {
						return err
					}
				}
				abiABI, err := abi.JSON(strings.NewReader(string(abiData)))
				if err != nil {
					return err
				}
				for _, error := range abiABI.Errors {
					abiErrors = append(abiErrors, error)
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return NewErrorsRepository(abiErrors)
}

func NewErrorsRepository(customErrors []abi.Error) (*ErrorsRepository, error) {
	defaultErrs, err := defaultErrors()
	if err != nil {
		return nil, err
	}
	customErrors = append(customErrors, defaultErrs...)
	er := ErrorsRepository{
		errs: make(map[[4]byte]abi.Error),
	}
	for _, e := range customErrors {
		e := abi.NewError(e.Name, e.Inputs)
		if err := er.Add(e); err != nil {
			return nil, err
		}
	}
	return &er, nil
}

func (r *ErrorsRepository) Add(e abi.Error) error {
	var sel [4]byte
	copy(sel[:], e.ID[:4])
	if _, ok := r.errs[sel]; ok {
		return fmt.Errorf("duplicate error selector: error=%v sel=%x", e.String(), sel)
	}
	r.errs[sel] = e
	return nil
}

func (r *ErrorsRepository) GetError(sel [4]byte) (abi.Error, bool) {
	e, ok := r.errs[sel]
	return e, ok
}

func (r *ErrorsRepository) ParseError(bz []byte) (string, interface{}, error) {
	if len(bz) < 4 {
		return "", nil, fmt.Errorf("invalid error data: %v", bz)
	}
	var sel [4]byte
	copy(sel[:], bz[:4])
	e, ok := r.GetError(sel)
	if !ok {
		return "", nil, fmt.Errorf("unknown error: sel=%x", sel)
	}
	v, err := e.Unpack(bz)
	return e.Sig, v, err
}

func defaultErrors() ([]abi.Error, error) {
	strT, err := abi.NewType("string", "", nil)
	if err != nil {
		return nil, err
	}
	uintT, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, err
	}

	errors := []abi.Error{
		{
			Name: "Error",
			Inputs: []abi.Argument{
				{
					Type: strT,
				},
			},
		},
		{
			Name: "Panic",
			Inputs: []abi.Argument{
				{
					Type: uintT,
				},
			},
		},
	}
	return errors, nil
}

func GetRevertReason(revertReason []byte, abiPaths []string) string {
	erepo, err := GetErepo(abiPaths)
	if err != nil {
		return fmt.Sprintf("GetErepo err=%v", err)
	}
	sig, args, err := erepo.ParseError(revertReason)
	if err != nil {
		return fmt.Sprintf("raw-revert-reason=\"%x\" parse-err=\"%v\"", revertReason, err)
	}
	return fmt.Sprintf("revert-reason=\"%v\" args=\"%v\"", sig, args)
}
