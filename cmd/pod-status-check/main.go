package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kuberhealthy/kuberhealthy/v3/pkg/checkclient"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	// Required for oidc kubectl testing.
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

// Options stores dependencies for the check.
type Options struct {
	// client is the Kubernetes client.
	client kubernetes.Interface
}

// init enables checkclient debug logging.
func init() {
	// Enable checkclient debug output for parity with v2 behavior.
	checkclient.Debug = true
}

// main loads configuration and runs the pod status check.
func main() {
	// Create the Kubernetes client.
	kubeConfigFile := os.Getenv("KUBECONFIG")
	client, err := createKubeClient(kubeConfigFile)
	if err != nil {
		log.Fatalln("Unable to create kubernetes client", err)
	}

	// Configure options.
	o := Options{client: client}

	// Find pods that are not running.
	failures, err := o.findPodsNotRunning(context.Background())
	if err != nil {
		reportFailureAndExit(err)
		return
	}

	// Report failures when any are found.
	if len(failures) >= 1 {
		log.Infoln("Amount of failures found:", len(failures))
		reportFailureListAndExit(failures)
		return
	}

	// Report success when no failures are found.
	err = checkclient.ReportSuccess()
	log.Infoln("Reporting Success, no unhealthy pods found.")
	if err != nil {
		log.Println("Error reporting success to Kuberhealthy servers", err)
		os.Exit(1)
	}
}

// findPodsNotRunning finds pods older than the skip duration in unhealthy phases.
func (o Options) findPodsNotRunning(ctx context.Context) ([]string, error) {
	// Initialize the failure list.
	var failures []string

	// Read environment configuration.
	skipDurationEnv := os.Getenv("SKIP_DURATION")
	namespace := os.Getenv("TARGET_NAMESPACE")
	if namespace == "" {
		log.Println("looking for pods across all namespaces, this requires a cluster role")
		namespace = v1.NamespaceAll
	}
	if namespace != v1.NamespaceAll {
		log.Printf("looking for pods in namespace %s", namespace)
	}

	// List pods excluding Kuberhealthy check pods.
	pods, err := o.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: "app!=kuberhealthy-check,source!=kuberhealthy"})
	if err != nil {
		return failures, err
	}

	// Parse the skip duration.
	skipDuration, err := time.ParseDuration(skipDurationEnv)
	if err != nil {
		log.Println("failed to parse skip duration:", err.Error())
		reportFailureAndExit(fmt.Errorf("failed to parse skip duration: %w", err))
		return failures, err
	}

	// Calculate the threshold time for skipping pods.
	checkTime := time.Now()
	skipBarrier := checkTime.Add(-skipDuration)

	// Start iteration over pods.
	for _, pod := range pods.Items {
		// Skip pods that are too young.
		if pod.CreationTimestamp.Time.After(skipBarrier) {
			log.Println("skipping checks on pod because it is too young:", pod.Name)
			continue
		}

		// Record unhealthy pod phases.
		switch {
		case pod.Status.Phase == v1.PodRunning:
			continue
		case pod.Status.Phase == v1.PodSucceeded:
			continue
		case pod.Status.Phase == v1.PodPending:
			failures = append(failures, "pod: "+pod.Name+" in namespace: "+pod.Namespace+" is in pod status phase "+string(pod.Status.Phase)+" ")
		case pod.Status.Phase == v1.PodFailed:
			failures = append(failures, "pod: "+pod.Name+" in namespace: "+pod.Namespace+" is in pod status phase "+string(pod.Status.Phase)+" ")
		case pod.Status.Phase == v1.PodUnknown:
			failures = append(failures, "pod: "+pod.Name+" in namespace: "+pod.Namespace+" is in pod status phase "+string(pod.Status.Phase)+" ")
		default:
			log.Info("pod: " + pod.Name + " in namespace: " + pod.Namespace + " is not in one of the five possible pod status phases " + string(pod.Status.Phase) + " ")
		}
	}

	return failures, nil
}

// reportFailureAndExit reports a failure and exits the process.
func reportFailureAndExit(err error) {
	// Report the failure to Kuberhealthy.
	reportErr := checkclient.ReportFailure([]string{err.Error()})
	if reportErr != nil {
		log.Println("Error", reportErr)
		os.Exit(1)
	}

	// Exit after reporting failure.
	os.Exit(1)
}

// reportFailureListAndExit reports multiple failures and exits.
func reportFailureListAndExit(failures []string) {
	// Report the failures to Kuberhealthy.
	err := checkclient.ReportFailure(failures)
	if err != nil {
		log.Println("Error reporting failures to Kuberhealthy servers", err)
		os.Exit(1)
	}

	// Exit after reporting failures.
	os.Exit(1)
}
