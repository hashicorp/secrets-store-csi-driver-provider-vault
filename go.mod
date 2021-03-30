module github.com/hashicorp/vault-csi-provider

go 1.12

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/hashicorp/go-hclog v0.8.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200605160147-a5ece683394c // indirect
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/client-go v0.20.0
	sigs.k8s.io/secrets-store-csi-driver v0.0.20
)
