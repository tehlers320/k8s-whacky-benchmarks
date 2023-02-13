package config

import (
	"bytes"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"github.com/cenkalti/backoff/v4"
)



var (
	defaultConfig = []byte(`
logging.level: "info"
test:
  deployment: 
    namespace: fortio
    name: fortioserver
  verticalScaleIncrease: 1
  delayBetweenTests: "5m"
  cpu:
    increaseAmount: 100
    max: 96000
  memory:
    # 100m 104857600
    increaseAmount: 104857600
    max: 104857600000
fortio:
  url: "http://localhost:8080/fortio/"
  duration: 10s
  repeattest: 3
  runsWithoutImprovement: 10
tests_to_run_from_configmaps:
- uno
- dos
- tres
- cuatro
- cinco
- seis
- siete
- ocho
- nueve
- diez
`)

)

func InitConfig() error {
	var err error
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("KWB")
	var cfg []byte = nil
	cfg = defaultConfig
	viper.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
	err = viper.ReadConfig(bytes.NewBuffer(cfg))
	if err != nil {
		return err
	}
	return err
}

type K8s struct {
	Local bool
}

func (k8s *K8s) CreateClient() (*kubernetes.Clientset, error) {
	config, err := k8s.buildConfig()

	if err != nil {
		return nil, errors.Wrapf(err, "error setting up cluster config")
	}

	return kubernetes.NewForConfig(config)
}

func (k8s *K8s) buildConfig() (*rest.Config, error) {
	if k8s.Local {
		log.Debug("Using local kubeconfig.")
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	log.Debug("Using in cluster kubeconfig.")
	return rest.InClusterConfig()
}

func NewExponentialBackOff() *backoff.ExponentialBackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 10 * time.Second
	b.Multiplier = 1.0
	b.MaxInterval = 60 * time.Second
	b.MaxElapsedTime = time.Minute * 1
	b.Reset()
	return b
}
