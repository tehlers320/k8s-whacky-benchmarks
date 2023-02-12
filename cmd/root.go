package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	config "github.com/tehlers320/k8s-whacky-benchmarks/config"
	fc "github.com/tehlers320/k8s-whacky-benchmarks/fortio"
	kates "github.com/tehlers320/k8s-whacky-benchmarks/k8s"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "generated code example",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//      Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	meta := &fc.Metadata{
		URL:              "http://localhost:8080/echo",
		Connections:      "1",
		Numcalls:         "10",
		Async:            "on",
		Save:             "off",
	}
	testFortioConnectivity := &fc.FortioRest{
		Metadata: *meta,
	}
	config.InitConfig()
	runId := fc.StartRun(testFortioConnectivity)
	if runId.RunID == 0 {
		panic("error fortio not responding")
	}
	k8s := config.K8s{Local: true}
	k8sClient, err := k8s.CreateClient()
	kClient := kates.NewK8sClient(k8sClient) 
	kClient.StartTests()
	if err != nil {
		log.Error(err, "unable to start k8s client")
		os.Exit(1)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9090", nil)
}

func init() {
	cobra.OnInitialize()
	switch os.Getenv("LOGGING_LEVEL") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}

}
