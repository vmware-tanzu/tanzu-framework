package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// State is client state stored locally.
type State map[string]interface{}

const (
	// StateFileName is the name of the state file.
	StateFileName = "state.yaml"
)

// GetState values.
func GetState() (state State, err error) {
	stateFilePath, err := stateFilePath()
	if err != nil {
		return state, err
	}
	b, err := ioutil.ReadFile(stateFilePath)
	if err != nil {
		return state, err
	}
	err = yaml.Unmarshal(b, &state)
	if err != nil {
		return state, errors.Wrap(err, "could not unmarshal state file")
	}
	return
}

// SetState values.
func SetState(vals map[string]interface{}) error {
	state, err := GetState()
	if os.IsNotExist(err) {
		state = map[string]interface{}{}
	} else if err != nil {
		return err
	}
	for k, v := range vals {
		state[k] = v
	}
	// TODO (pbarker): handle potential races
	return state.store()
}

// DeleteState keys.
func DeleteState(keys ...string) error {
	state, err := GetState()
	if os.IsNotExist(err) {
		return fmt.Errorf("state file does not exist")
	}
	for _, key := range keys {
		delete(state, key)
	}
	return state.store()
}

func (s State) store() error {
	localDir, err := LocalDir()
	if err != nil {
		return err
	}
	err = ensurePath(localDir)
	if err != nil {
		return err
	}
	stateFilePath, err := stateFilePath()
	if err != nil {
		return err
	}
	b, err := yaml.Marshal(s)
	if err != nil {
		return errors.Wrap(err, "could not marshal state file")
	}
	if err = ioutil.WriteFile(stateFilePath, b, 0644); err != nil {
		return errors.Wrap(err, "could dnot write state file")
	}
	return nil
}

func stateFilePath() (path string, err error) {
	localDir, err := LocalDir()
	if err != nil {
		return path, err
	}
	path = filepath.Join(localDir, StateFileName)
	return
}

func ensurePath(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
		return nil
	}
	return err
}
