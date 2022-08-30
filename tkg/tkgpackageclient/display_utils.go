// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/mattn/go-isatty"

	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

// DisplayProgress creates an spinner instance; keeps receiving the progress messages in the channel and displays those using the spinner until an error occurs
func DisplayProgress(initialMsg string, pp *tkgpackagedatamodel.PackageProgress) error { //nolint:gocyclo
	var (
		currMsg string
		s       *spinner.Spinner
		err     error
	)

	newSpinner := func() (*spinner.Spinner, error) {
		s = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		if err := s.Color("bgBlack", "bold", "fgWhite"); err != nil {
			return nil, err
		}
		return s, nil
	}
	if s, err = newSpinner(); err != nil {
		return err
	}

	s.Suffix = fmt.Sprintf(" %s", initialMsg)
	// Start the spinner only if attached to terminal
	attachedToTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	if attachedToTerminal {
		s.Start()
	}

	writeProgressToSpinner := func(s *spinner.Spinner, msg string) error {
		if !attachedToTerminal {
			return nil
		}
		spinnerMsg := s.Suffix
		s.Stop()
		if s, err = newSpinner(); err != nil {
			return err
		}
		// Stopping the spinner would eliminate the whole line
		// We want to keep the spinner message to give users more contexts
		log.Infof("%s\n", spinnerMsg)
		s.Suffix = fmt.Sprintf(" %s", msg)
		s.Start()
		return nil
	}

	defer func() {
		if attachedToTerminal {
			s.Stop()
		}
	}()
	for {
		select {
		case err := <-pp.Err:
			s.FinalMSG = "\n\n"
			return err
		case msg := <-pp.ProgressMsg:
			if msg != currMsg {
				if err := writeProgressToSpinner(s, msg); err != nil {
					return err
				}
				currMsg = msg
			}
		case <-pp.Done:
			for msg := range pp.ProgressMsg {
				if msg == currMsg {
					continue
				}
				if err := writeProgressToSpinner(s, msg); err != nil {
					return err
				}
				currMsg = msg
			}
			return nil
		}
	}
}

func newPackageProgress() *tkgpackagedatamodel.PackageProgress {
	return &tkgpackagedatamodel.PackageProgress{
		ProgressMsg: make(chan string, 10),
		Err:         make(chan error),
		Done:        make(chan struct{}),
	}
}
