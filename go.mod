module github.com/Octops/agones-event-broadcaster

go 1.16

require (
	agones.dev/agones v1.19.0
	cloud.google.com/go/pubsub v1.2.0
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/api v0.20.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.3
)
