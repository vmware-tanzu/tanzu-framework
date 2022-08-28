package lib

import (
	"bytes"
	"fmt"
	"time"

	clitest "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cli"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

type SecretPluginBase interface {
	AddRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult
	UpdateRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult
	DeleteRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult
	ListRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult
}

type SecretPlugin interface {
	// Base commands are implemented by the below interface
	SecretPluginBase
}

type SecretPluginResult struct {
	Pass   bool
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
	Error  error
}

type secretPlugin struct {
	kubeConfigPath string
	interval       time.Duration
	timeout        time.Duration
	outputFormat   string
	logFile        string
	verbose        int32
}

func NewSecretPlugin(kubeConfigPath string,
	interval time.Duration,
	timeout time.Duration,
	outputFormat string,
	verbose int32) SecretPlugin {
	return &secretPlugin{
		kubeConfigPath: kubeConfigPath,
		interval:       interval,
		timeout:        timeout,
		outputFormat:   outputFormat,
		verbose:        verbose,
	}
}

func (p *secretPlugin) addKubeConfig(cmd string) string {
	if cmd != "" && p.kubeConfigPath != "" {
		cmd += fmt.Sprintf(" --kubeconfig %s", p.kubeConfigPath)
	}
	return cmd
}

func (p *secretPlugin) addOutputFormat(cmd string) string {
	if cmd != "" && p.outputFormat != "" {
		cmd += fmt.Sprintf(" --output %s", p.outputFormat)
	}
	return cmd
}

func (p *secretPlugin) addGlobalOptions(cmd string) string {
	cmd = p.addVerbose(cmd)
	return cmd
}

func (p *secretPlugin) addVerbose(cmd string) string {
	if cmd != "" && p.verbose != 0 {
		cmd += fmt.Sprintf(" --verbose %d", p.verbose)
	}
	return cmd
}

func (p *secretPlugin) AddRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult {
	var result SecretPluginResult
	cmd := fmt.Sprintf("tanzu secret registry add %s --server %s --username %s", o.SecretName, o.Server, o.Username)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.PasswordInput != "" {
		cmd += fmt.Sprintf(" --password %s", o.PasswordInput)
	}
	if o.PasswordFile != "" {
		cmd += fmt.Sprintf(" --password-file %s", o.PasswordFile)
	}
	if o.PasswordEnvVar != "" {
		cmd += fmt.Sprintf(" --password-env-var %s", o.PasswordEnvVar)
	}
	if o.ExportToAllNamespaces {
		cmd += fmt.Sprintf(" --export-to-all-namespaces")
	}
	if o.SkipPrompt {
		cmd += fmt.Sprintf(" -y")
	}

	cmd = p.addKubeConfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *secretPlugin) DeleteRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult {
	var result SecretPluginResult
	cmd := fmt.Sprintf("tanzu secret registry delete %s", o.SecretName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.SkipPrompt {
		cmd += fmt.Sprintf(" -y")
	}

	cmd = p.addKubeConfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *secretPlugin) ListRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult {
	var result SecretPluginResult
	cmd := fmt.Sprintf("tanzu secret registry list")
	if o.AllNamespaces {
		cmd += fmt.Sprintf(" --all-namespaces")
	}
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}

	cmd = p.addOutputFormat(cmd)
	cmd = p.addKubeConfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}

func (p *secretPlugin) UpdateRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) SecretPluginResult {
	var result SecretPluginResult
	cmd := fmt.Sprintf("tanzu secret registry update %s", o.SecretName)
	if o.Namespace != "" {
		cmd += fmt.Sprintf(" --namespace %s", o.Namespace)
	}
	if o.Username != "" {
		cmd += fmt.Sprintf(" --username %s", o.Username)
	}
	if o.PasswordInput != "" {
		cmd += fmt.Sprintf(" --password %s", o.PasswordInput)
	}
	if o.PasswordFile != "" {
		cmd += fmt.Sprintf(" --password-file %s", o.PasswordFile)
	}
	if o.PasswordEnvVar != "" {
		cmd += fmt.Sprintf(" --password-env-var %s", o.PasswordEnvVar)
	}
	if o.Export.ExportToAllNamespaces != nil {
		if *o.Export.ExportToAllNamespaces == true {
			cmd += fmt.Sprintf(" --export-to-all-namespaces=true")
		} else {
			cmd += fmt.Sprintf(" --export-to-all-namespaces=false")
		}
	}
	if o.SkipPrompt {
		cmd += fmt.Sprintf(" -y")
	}

	cmd = p.addKubeConfig(cmd)
	cmd = p.addGlobalOptions(cmd)
	result.Stdout, result.Stderr, result.Error = clitest.Exec(cmd)
	return result
}
