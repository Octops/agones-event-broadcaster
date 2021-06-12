module github.com/Octops/agones-event-broadcaster

go 1.16

require (
	agones.dev/agones v1.11.0
	cloud.google.com/go/pubsub v1.0.1
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.5.0
	google.golang.org/api v0.13.0
	k8s.io/api v0.17.14
	k8s.io/apimachinery v0.17.14
	k8s.io/client-go v0.17.14
	sigs.k8s.io/controller-runtime v0.5.4
)
