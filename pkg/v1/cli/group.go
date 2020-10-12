package cli

// cmdGroup is a group of CLI commands.
type cmdGroup string

const (
	// RunCmdGroup are commands associated with Tanzu Run.
	RunCmdGroup cmdGroup = "Run"

	// ManageCmdGroup are commands associated with Tanzu Manage.
	ManageCmdGroup cmdGroup = "Manage"

	// BuildCmdGroup are commands associated with Tanzu Build.
	BuildCmdGroup cmdGroup = "Build"

	// ObserveCmdGroup are commands associated with Tanzu Observe.
	ObserveCmdGroup cmdGroup = "Observe"

	// SystemCmdGroup are system commands.
	SystemCmdGroup cmdGroup = "System"

	// VersionCmdGroup are version commands.
	VersionCmdGroup cmdGroup = "Version"
)
