module github.com/Octops/gameserver-events-broadcaster

go 1.14

require (
	agones.dev/agones v1.5.0
	cloud.google.com/go/pubsub v1.0.1
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.3.0
	k8s.io/api v0.0.0-20191004102349-159aefb8556b
	k8s.io/apimachinery v0.0.0-20191004074956-c5d2f014d689
	k8s.io/client-go v11.0.1-0.20191029005444-8e4128053008+incompatible
	sigs.k8s.io/controller-runtime v0.0.0-00010101000000-000000000000
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.3.0
)
