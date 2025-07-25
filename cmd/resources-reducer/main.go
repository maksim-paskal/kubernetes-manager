package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var kubeconfig = flag.String("kubeconfig", os.Getenv("KUBECONFIG"), "absolute path to the kubeconfig file")
var hpaDisabled = flag.Bool("hpa-disabled", true, "disable HPA")
var pdbDelete = flag.Bool("pdb-delete", true, "delete PDBs")
var scaleDeployment = flag.Bool("scale-deployment", true, "scale deployments to 1 replica")
var minimumAge = flag.Duration("minimum-age", 10*time.Minute, "minimum age of object to process")

func start(ctx context.Context) error {
	slog.Info("Use kubeconfig file", "path", *kubeconfig)

	restconfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return errors.Wrap(err, "error in clientcmd.BuildConfigFromFlags")
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		return errors.Wrap(err, "error in kubernetes.NewForConfig")
	}

	slog.Info("Get all namespaces")

	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: "environment=dev,kubernetes-manager=true",
	})
	if err != nil {
		return errors.Wrap(err, "error in clientset.CoreV1().Namespaces().List")
	}

	for _, ns := range namespaces.Items {
		slog := slog.With("namespace", ns.Name)

		slog.Info("Namespace found")

		if _, ok := ns.Annotations["region"]; ok {
			slog.Info("Removing region annotation", "annotation", ns.Annotations["region"])

			payload := `[{"op": "remove", "path": "/metadata/annotations/region"}]`
			if _, err := clientset.CoreV1().Namespaces().Patch(ctx, ns.Name,
				types.JSONPatchType,
				[]byte(payload),
				metav1.PatchOptions{},
			); err != nil {
				return errors.Wrapf(err, "error patching namespace %s", ns.Name)
			}
		}
		var wg sync.WaitGroup

		funcs := []func(){}

		funcs = append(funcs, func() {
			defer wg.Done()

			if !*pdbDelete {
				slog.Info("Skipping PDB deletion as per flag")
				return
			}

			pdbs, err := clientset.PolicyV1().PodDisruptionBudgets(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				slog.Error("Error listing PDBs", "error", err)
				return
			}
			for _, pdb := range pdbs.Items {
				slog := slog.With("pdb", pdb.Name)
				slog.Info("Deleting PodDisruptionBudget")
				if err := clientset.PolicyV1().PodDisruptionBudgets(ns.Name).Delete(ctx, pdb.Name, metav1.DeleteOptions{}); err != nil {
					slog.Error("Error deleting PodDisruptionBudget", "error", err)
				}
			}
		})

		funcs = append(funcs, func() {
			defer wg.Done()

			if !*scaleDeployment {
				slog.Info("Skipping deployment scaling as per flag")
				return
			}

			deployments, err := clientset.AppsV1().Deployments(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				slog.Error("Error listing deployments", "error", err)
				return
			}
			for _, deployment := range deployments.Items {
				slog := slog.With("deployment", deployment.Name)

				if deployment.Spec.Replicas == nil || *deployment.Spec.Replicas <= 1 {
					continue
				}

				if time.Since(deployment.CreationTimestamp.Time) < *minimumAge {
					slog.Info("Skipping Deployment modification as it is too young")
					continue
				}

				slog.Info("Scaling deployment to 1 replica", "currentReplicas", *deployment.Spec.Replicas)

				deployment.Spec.Replicas = new(int32)
				*deployment.Spec.Replicas = 1

				_, err := clientset.AppsV1().Deployments(ns.Name).Update(ctx, &deployment, metav1.UpdateOptions{})
				if err != nil {
					slog.Error("Error scaling deployment", "error", err)
				}
			}
		})

		funcs = append(funcs, func() {
			defer wg.Done()

			const disabledPrefix = "disabled-"

			hpas, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				slog.Error("Error listing HPAs", "error", err)
				return
			}

			for _, hpa := range hpas.Items {
				slog := slog.With("hpa", hpa.Name)

				if time.Since(hpa.CreationTimestamp.Time) < *minimumAge {
					slog.Info("Skipping HPA modification as it is too young")
					continue
				}

				scaleTargetRefKind := strings.TrimPrefix(hpa.Spec.ScaleTargetRef.Kind, disabledPrefix)

				if *hpaDisabled {
					scaleTargetRefKind = disabledPrefix + scaleTargetRefKind
				}

				// if already set - skip
				if hpa.Spec.ScaleTargetRef.Kind == scaleTargetRefKind {
					continue
				}

				slog.Info("Disabling HPA", "kind", hpa.Spec.ScaleTargetRef.Kind)

				payload := fmt.Sprintf(`{"spec":{"scaleTargetRef":{"kind": "%s" }}}`, scaleTargetRefKind)

				_, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(ns.Name).Patch(ctx,
					hpa.Name,
					types.StrategicMergePatchType,
					[]byte(payload),
					metav1.PatchOptions{},
				)
				if err != nil {
					slog.Error("Error patching HPA", "error", err)
				}
			}
		})

		funcs = append(funcs, func() {
			defer wg.Done()

			pods, err := clientset.CoreV1().Pods(ns.Name).List(ctx, metav1.ListOptions{})
			if err != nil {
				slog.Error("Error listing pods", "error", err)
				return
			}

			for _, pod := range pods.Items {
				slog := slog.With("pod", pod.Name)
				if pod.Status.Phase == "Running" {
					continue
				}

				if time.Since(pod.CreationTimestamp.Time) < *minimumAge {
					slog.Info("Skipping Pod deletion as it is too young")
					continue
				}
				slog.Info("Delete Pod")
				if err := clientset.CoreV1().Pods(ns.Name).Delete(ctx, pod.Name, metav1.DeleteOptions{}); err != nil {
					slog.Error("Error deleting Pod", "error", err)
				}
			}
		})

		for _, fn := range funcs {
			wg.Add(1)
			go fn()
		}

		wg.Wait()
	}
	return nil
}

func main() {
	if err := start(context.Background()); err != nil {
		log.Fatal(err)
		return
	}
}
