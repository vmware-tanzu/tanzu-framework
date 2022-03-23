# telemetry
## Summary
Plugin for configuring cluster-wide telemetry settings. These settings are applicable to any generic kubernetes cluster.
Settings are respected by any tanzu product exercising the telemetry SDK.

## Usage

### status
#### help output
```shell
$ tanzu telemetry status --help
Status of tanzu telemetry settings

Usage:
  tanzu telemetry status [flags]

Examples:

    # get status
    tanzu telemetry status

Flags:
  -h, --help   help for status
```

#### printing status of cluster
```shell
$ tanzu telemetry status
- ceip: |
    level: disabled
  shared_identifiers: |
    customer_csp_org_id: some-org
    customer_entitlement_account_number: ABCD
    env_is_prod: "true"
```

### update
#### help output
```shell
$ tanzu telemetry update --help
Update tanzu telemetry settings

Usage:
  tanzu telemetry update [flags]

Examples:

    # opt into ceip
    tanzu telemetry update --ceip-opt-in
	# opt out of ceip
    tanzu telemetry update --ceip-opt-out
	# update shared configuration settings
    tanzu telemetry update --env-is-prod "true" --entitlement-account-number "1234" --csp-org-id "XXXX"


Flags:
      --ceip-opt-in                         opt into VMware's ceip program
      --ceip-opt-out                        opt out of VMware's ceip program
      --csp-org-id string                   Accepts a string and sets a cluster-wide CSP
                                                                            org ID. Empty string is equivalent to
                                                                            unsetting this value.
      --entitlement-account-number string   Accepts a string and sets a cluster-wide
                                                                            entitlement account number. Empty string is
                                                                            equivalent to unsetting this value
      --env-is-prod string                  Accepts a boolean and sets a cluster-wide
                                                                            value denoting whether the target is a
                                                                            production cluster or not.
  -h, --help                                help for update
```

#### opt in/out of telemetry
```shell
$ tanzu telemetry update --ceip-opt-in

$ tanzu telemetry update --ceip-opt-out

```

#### update shared identifiers
```shell
$ tanzu telemetry update --csp-org-id "test-org" --entitlement-account-number "XXXX" --env-is-prod false
found existing identifiers config map
Updating config map ....
```