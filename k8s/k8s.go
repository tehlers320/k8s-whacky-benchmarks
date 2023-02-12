package k8sattacks

import (
	"context"
	_ "embed"
	"time"

	"github.com/spf13/viper"

	"k8s.io/apimachinery/pkg/api/resource"

	log "github.com/sirupsen/logrus"
	config "github.com/tehlers320/k8s-whacky-benchmarks/config"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fc "github.com/tehlers320/k8s-whacky-benchmarks/fortio"
	"k8s.io/client-go/kubernetes"
	"github.com/cenkalti/backoff/v4"
)

type k8s struct {
	k8sClient *kubernetes.Clientset
	namespace string
	deploymentName string
	bestQPS float64
	bestRunID string
	deployError bool
	runNumber int
	runsWithoutImprovement int
	currMemRq *resource.Quantity
	currMemLm *resource.Quantity
	currCpuRq *resource.Quantity
	currCpuLm *resource.Quantity
	deployStateHealthy bool
}

func NewK8sClient(client *kubernetes.Clientset) *k8s{
	return &k8s{k8sClient: client}

}

type K8s interface {
	StartTests() 
	CheckDeploymentState() bool
}

func (k *k8s)StartTests()  {
	Result := runFortio()
	if compareResult(k.bestQPS, Result.ActualQPS) {
		k.bestQPS = Result.ActualQPS
		k.bestRunID = Result.ResultID
	}
	k.runNumber = 0
	k.runsWithoutImprovement = 0
	k.verticallyScale()
	for {
		if k.deployError {
			log.Errorf("Deployment wont stabalize, shutting down")
			break
		}
		Result := runFortio()
		if compareResult(k.bestQPS, Result.ActualQPS) {
			k.bestQPS = Result.ActualQPS
			k.bestRunID = Result.ResultID
		}
		if !compareResult(k.bestQPS, Result.ActualQPS) {
			k.runsWithoutImprovement += 1
			if k.runsWithoutImprovement / viper.GetInt("fortio.repeattest") >= viper.GetInt("fortio.runsWithoutImprovement"){
				log.Info("Reached maximum runs without an improvement")
				break
			}
		}
		if !k.checkDeploymentState() {
			k.fixmemoryCrash()
			if !k.checkDeploymentState() {
				t := time.Now()
				if !k.backoffOnDeploy(t) {
					k.deployError = true
				}
			}
		}
		k.runNumber += 1
		if k.runNumber >= viper.GetInt("fortio.repeattest") {
			k.verticallyScale()
			k.runNumber = 0
		}
	}
}

func compareResult(previous float64, current float64) bool {
	return current > previous
}

func runFortio() *fc.RunResponse{
	meta := &fc.Metadata{
		URL:              "http://localhost:8080/echo",
		Connections:      "1",
		Async:            "off",
		Save:             "on",
		Qps:              "-1",
		DurStr:           viper.GetString("fortio.duration"),
	}
	testFortioConnectivity := &fc.FortioRest{
		Metadata: *meta,
	}
	RunResponse := fc.StartRun(testFortioConnectivity)
	log.Debugf("fortio run started %v", RunResponse)
	return RunResponse
}

func (k *k8s)fixmemoryCrash()  {
	MemoryIncrease := *resource.NewQuantity(viper.GetInt64("test.memory.increaseamount"), resource.BinarySI)
	k.currMemRq.Add(MemoryIncrease)
	k.currMemLm.Add(MemoryIncrease)
	d, err := k.k8sClient.AppsV1().
	Deployments(k.namespace).
	Get(context.TODO(), k.deploymentName, v1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	d.Spec.Template.Spec.Containers[0].Resources.Limits["memory"] = *k.currMemRq
	d.Spec.Template.Spec.Containers[0].Resources.Requests["memory"] = *k.currMemLm

	_, err = k.k8sClient.AppsV1().
	Deployments(k.namespace).
	Update(
		context.TODO(),
		d,
		v1.UpdateOptions{},
	)

	if err != nil {
		log.Fatal(err)
	}

	if !k.checkDeploymentState() {
		t := time.Now()
		if !k.backoffOnDeploy(t) {
			k.deployError = true
		}
	}

}


func (k *k8s)verticallyScale()  {
	CpuIncrease := *resource.NewQuantity(viper.GetInt64("test.cpu.increaseammount"), resource.DecimalSI)
	
	d, err := k.k8sClient.AppsV1().
	Deployments(k.namespace).
	Get(context.TODO(), k.deploymentName, v1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	// TODO: only does pod 1, whatevers...
	k.currCpuRq = d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu()
	k.currCpuLm = d.Spec.Template.Spec.Containers[0].Resources.Limits.Cpu()
	k.currCpuRq.Add(CpuIncrease)
	k.currCpuRq.Add(CpuIncrease)

	d.Spec.Template.Spec.Containers[0].Resources.Limits["cpu"] = *k.currCpuLm
	d.Spec.Template.Spec.Containers[0].Resources.Requests["cpu"] = *k.currCpuRq

    _, err = k.k8sClient.AppsV1().
        Deployments(k.namespace).
        Update(
			context.TODO(),
			d,
			v1.UpdateOptions{},
		)
    if err != nil {
        log.Fatal(err)
    }
	
	if !k.checkDeploymentState() {
		t := time.Now()
		k.backoffOnDeploy(t)
	}
	log.Infof("%s", &d)
}

func (k *k8s)backoffOnDeploy(t time.Time) bool{
	ExpBackOff := config.NewExponentialBackOff()
	operation := func() bool {
		log.Debugf("entering backoff retry for time: %v", t)

		return !k.checkDeploymentState()
	}

	ticker := backoff.NewTicker(ExpBackOff)

	for _ = range ticker.C {
		if ok := operation(); ok == false {
			log.Warnf("will retry timestamp %v in ... %v", t, ExpBackOff.NextBackOff())
			continue
		}

		ticker.Stop()
		break
	}

	return true
}

func (k *k8s)checkDeploymentState() bool {
	d, err := k.k8sClient.AppsV1().
	Deployments(k.namespace).
	Get(context.TODO(), k.deploymentName, v1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	if d != nil &&
		d.Spec.Replicas != nil &&
		*d.Spec.Replicas == d.Status.ReadyReplicas {
		log.Info("Deployment is READY")
		k.deployStateHealthy = true
		return true
	} else {
		log.Infof("Deployment is NOT READY, count: %d", *d.Spec.Replicas)
		k.deployStateHealthy = false
	}
	k.deployStateHealthy = true
	return false
}

func (k *k8s)scaleUpDeployment()  {
	s, err := k.k8sClient.AppsV1().
	Deployments(k.namespace).
	GetScale(context.TODO(), k.deploymentName, v1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}
	sc := *s
    sc.Spec.Replicas = 10

    pods, err := k.k8sClient.AppsV1().
        Deployments(k.namespace).
        UpdateScale(context.TODO(),
            k.deploymentName, &sc, v1.UpdateOptions{})
    if err != nil {
        log.Fatal(err)
    }
	log.Infof("%s", &pods.Status)
}
