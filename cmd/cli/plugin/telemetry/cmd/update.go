package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/telemetry/update"

	"github.com/spf13/pflag"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"k8s.io/client-go/dynamic"
)

var (
	ean, cspOrgId, envIsProd string
	ceipOptIn, ceipOptOut    bool
)

var identifierFlagMapping = map[string]string{
	UpdateEanFlag:       "customer_entitlement_account_number",
	UpdateCspOrgFlag:    "customer_csp_org_id",
	UpdateEnvIsProdFlag: "env_is_prod",
}

type UpdateCmd struct {
	Cmd          *cobra.Command
	ClientGetter func() (dynamic.Interface, error)
	out          component.OutputWriter
}

// NewUpdateCmd creates an update cmd and injects a function for retrieving a k8s client
// Allows us to unit test by injecting a fake client
func NewUpdateCmd(clientGetter func() (dynamic.Interface, error), out component.OutputWriter) *UpdateCmd {
	out.SetKeys("CEIP", "SHARED IDENTIFIERS")

	uc := UpdateCmd{
		ClientGetter: clientGetter,
		out:          out,
	}
	uc.Cmd = &cobra.Command{
		Use:   "update",
		Short: "Update tanzu telemetry settings",
		Example: `
    # opt into ceip
    tanzu telemetry update --ceip-opt-in
	# opt out of ceip
    tanzu telemetry update --ceip-opt-out
	# update shared configuration settings
    tanzu telemetry update --env-is-prod "true" --entitlement-account-number "1234" --csp-org-id "XXXX"
`,
		RunE: uc.Update,
	}

	uc.Cmd.Flags().BoolVar(&ceipOptIn, UpdateCeipOptInFlag, false, "opt into VMware's ceip program")
	uc.Cmd.Flags().BoolVar(&ceipOptOut, UpdateCeipOptOutFlag, false, "opt out of VMware's ceip program")
	uc.Cmd.Flags().StringVar(&ean, UpdateEanFlag, "", `Accepts a string and sets a cluster-wide
                                entitlement account number. Empty string is
                                equivalent to unsetting this value`)
	uc.Cmd.Flags().StringVar(&cspOrgId, UpdateCspOrgFlag, "", `Accepts a string and sets a cluster-wide CSP
                                org ID. Empty string is equivalent to
                                unsetting this value.`)
	uc.Cmd.Flags().StringVar(&envIsProd, UpdateEnvIsProdFlag, "", `Accepts a boolean and sets a cluster-wide
                                value denoting whether the target is a
                                production cluster or not.`)

	return &uc
}

// Update configures telemetry settings on the targeted cluster
func (uc *UpdateCmd) Update(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	anyFlagSet := false
	uc.Cmd.Flags().Visit(func(f *pflag.Flag) {
		anyFlagSet = true
	})
	if !anyFlagSet {
		return errors.New("must set at least one flag")
	}

	if uc.Cmd.Flags().Changed(UpdateCeipOptInFlag) && uc.Cmd.Flags().Changed(UpdateCeipOptOutFlag) {
		return errors.New("cannot set both ceip-opt-in and ceip-opt-out flags")
	}

	dynamicClient, err := uc.ClientGetter()
	if err != nil {
		return err
	}

	updater := update.Updater{Client: dynamicClient}

	if uc.Cmd.Flags().Changed(UpdateCeipOptInFlag) || uc.Cmd.Flags().Changed(UpdateCeipOptOutFlag) {
		var optInVal bool
		if ceipOptIn {
			optInVal = true
		} else if ceipOptOut {
			optInVal = false
		}
		err := updater.UpdateCeip(ctx, optInVal)
		if err != nil {
			return err
		}
	}

	if uc.Cmd.Flags().Changed(UpdateEanFlag) || uc.Cmd.Flags().Changed(UpdateCspOrgFlag) || uc.Cmd.Flags().Changed(UpdateEnvIsProdFlag) {
		valsToUpdate, err := uc.buildValsToUpdate()
		if err != nil {
			return err
		}
		err = updater.UpdateIdentifiers(ctx, valsToUpdate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (uc *UpdateCmd) buildValsToUpdate() ([]update.UpdateVal, error) {
	var valsToUpdate []update.UpdateVal

	for idFlag, keyName := range identifierFlagMapping {
		val, err := uc.Cmd.Flags().GetString(idFlag)
		if err != nil {
			return nil, err
		}
		if idFlag == UpdateEnvIsProdFlag {
			val = strings.ToLower(val)
		}
		uVal := update.UpdateVal{
			Changed: uc.Cmd.Flags().Changed(idFlag),
			Key:     keyName,
			Value:   val,
		}

		valsToUpdate = append(valsToUpdate, uVal)
	}

	return valsToUpdate, nil
}
