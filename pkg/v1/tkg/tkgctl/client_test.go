// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"io/ioutil"
	"os"
	"sync"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/providerinterface"
)

var _ = Describe("ensurePrerequisite", func() {
	var (
		err            error
		providerGetter providerinterface.ProviderInterface
	)

	BeforeEach(func() {
		testingDir, err = ioutil.TempDir("", "test")
		err = os.MkdirAll(testingDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
		providerGetter = getDefaultProviderGetter()
	})

	Context("When two goroutines try to modify the file under configDir", func() {
		It("should not return errors", func() {
			errs := make(chan error, 2)
			defer close(errs)
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				err := ensurePrerequisite(testingDir, providerGetter)
				errs <- err
			}()

			go func() {
				defer wg.Done()
				err := ensurePrerequisite(testingDir, providerGetter)
				errs <- err
			}()
			wg.Wait()
			var err1, err2 error
			err1 = <-errs
			err2 = <-errs

			Expect(err1).ToNot(HaveOccurred())
			Expect(err2).ToNot(HaveOccurred())
		})
	})
})
