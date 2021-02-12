package config

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
)

const (
	defaultVaultAddress                 string = "https://127.0.0.1:8200"
	defaultKubernetesServiceAccountPath string = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	defaultVaultKubernetesMountPath     string = "kubernetes"
)

// Config represents all of the provider's configurable behaviour from the MountRequest proto message:
// * Parameters from the `Attributes` field.
// * Plus the rest of the proto fields we consume.
// See sigs.k8s.io/secrets-store-csi-driver/provider/v1alpha1/service.pb.go
type Config struct {
	Parameters
	TargetPath     string
	FilePermission os.FileMode
}

// Parameters stores the parameters specified in a mount request's `Attributes` field.
// It consists of the parameters section from the SecretProviderClass being mounted
// and pod metadata provided by the driver.
//
// Top-level values that aren't strings are not directly deserialisable because
// they are defined as literal string types:
// https://github.com/kubernetes-sigs/secrets-store-csi-driver/blob/0ba9810d41cc2dc336c68251d45ebac19f2e7f28/apis/v1alpha1/secretproviderclass_types.go#L59
//
// So we just deserialise by hand to avoid complexity and two passes.
type Parameters struct {
	VaultRoleName                string
	VaultAddress                 string
	VaultKubernetesMountPath     string
	KubernetesServiceAccountPath string
	TLSConfig                    TLSConfig
	Secrets                      []Secret
	PodInfo                      PodInfo
}

type TLSConfig struct {
	VaultCAPEM         string
	VaultCACertPath    string
	VaultCADirectory   string
	VaultTLSServerName string
	VaultSkipTLSVerify bool
}

type PodInfo struct {
	Name               string
	UID                types.UID
	Namespace          string
	ServiceAccountName string
}

func (c TLSConfig) CertificatesConfigured() bool {
	return c.VaultCAPEM != "" ||
		c.VaultCACertPath != "" ||
		c.VaultCADirectory != ""
}

type Secret struct {
	ObjectName string                 `yaml:"objectName,omitempty"`
	SecretPath string                 `yaml:"secretPath,omitempty"`
	SecretKey  string                 `yaml:"secretKey,omitempty"`
	Method     string                 `yaml:"method,omitempty"`
	SecretArgs map[string]interface{} `yaml:"secretArgs,omitempty"`
}

func Parse(parametersStr, targetPath, permissionStr string) (Config, error) {
	config := Config{
		TargetPath: targetPath,
	}

	var err error
	config.Parameters, err = parseParameters(parametersStr)
	if err != nil {
		return Config{}, err
	}

	err = json.Unmarshal([]byte(permissionStr), &config.FilePermission)
	if err != nil {
		return Config{}, err
	}

	err = config.Validate()
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func parseParameters(parametersStr string) (Parameters, error) {
	var params map[string]string
	err := json.Unmarshal([]byte(parametersStr), &params)
	if err != nil {
		return Parameters{}, err
	}

	var parameters Parameters
	parameters.VaultRoleName = params["roleName"]
	parameters.VaultAddress = params["vaultAddress"]
	parameters.TLSConfig.VaultCAPEM = params["vaultCAPem"]
	parameters.TLSConfig.VaultCACertPath = params["vaultCACertPath"]
	parameters.TLSConfig.VaultCADirectory = params["vaultCADirectory"]
	parameters.TLSConfig.VaultTLSServerName = params["vaultTLSServerName"]
	parameters.VaultKubernetesMountPath = params["vaultKubernetesMountPath"]
	parameters.KubernetesServiceAccountPath = params["KubernetesServiceAccountPath"]
	parameters.PodInfo.Name = params["csi.storage.k8s.io/pod.name"]
	parameters.PodInfo.UID = types.UID(params["csi.storage.k8s.io/pod.uid"])
	parameters.PodInfo.Namespace = params["csi.storage.k8s.io/pod.namespace"]
	parameters.PodInfo.ServiceAccountName = params["csi.storage.k8s.io/serviceAccount.name"]
	if skipTLS, ok := params["vaultSkipTLSVerify"]; ok {
		value, err := strconv.ParseBool(skipTLS)
		if err == nil {
			parameters.TLSConfig.VaultSkipTLSVerify = value
		} else {
			return Parameters{}, err
		}
	}

	secretsYaml := params["objects"]
	err = yaml.Unmarshal([]byte(secretsYaml), &parameters.Secrets)
	if err != nil {
		return Parameters{}, err
	}

	// Set default values.
	if parameters.VaultAddress == "" {
		parameters.VaultAddress = defaultVaultAddress
	}

	if parameters.VaultKubernetesMountPath == "" {
		parameters.VaultKubernetesMountPath = defaultVaultKubernetesMountPath
	}
	if parameters.KubernetesServiceAccountPath == "" {
		parameters.KubernetesServiceAccountPath = defaultKubernetesServiceAccountPath
	}

	return parameters, nil
}

func (c *Config) Validate() error {
	// Some basic validation checks.
	if c.TargetPath == "" {
		return errors.New("missing target path field")
	}
	if c.Parameters.VaultRoleName == "" {
		return errors.Errorf("missing 'roleName' in SecretProviderClass definition")
	}
	certificatesConfigured := c.Parameters.TLSConfig.CertificatesConfigured()
	if c.Parameters.TLSConfig.VaultSkipTLSVerify && certificatesConfigured {
		return errors.New("both vaultSkipTLSVerify and TLS configuration are set")
	}
	if !c.Parameters.TLSConfig.VaultSkipTLSVerify && !certificatesConfigured {
		return errors.New("no TLS configuration and vaultSkipTLSVerify is false, will use system CA certificates")
	}
	if len(c.Parameters.Secrets) == 0 {
		return errors.New("no secrets configured - the provider will not read any secret material")
	}

	return nil
}
