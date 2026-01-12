package main

import (
	"context"
	"os"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

// TestFindPodsNotRunning validates pod status detection in single and multi namespace modes.
func TestFindPodsNotRunning(t *testing.T) {
	// Build test fixtures.
	objects := getTestPods()
	os.Setenv("SKIP_DURATION", "10m")

	// Define test cases.
	type fields struct {
		namespace string
	}
	tests := []struct {
		name    string
		fields  fields
		want    []string
		wantErr bool
	}{
		{
			name: "singleNamespace",
			fields: fields{
				namespace: "foo",
			},
			want:    []string{"pod: foo-pod in namespace: foo is in pod status phase Pending "},
			wantErr: false,
		},
		{
			name: "multiNamespace",
			fields: fields{
				namespace: "",
			},
			want:    []string{"pod: bar-pod in namespace: bar is in pod status phase Pending ", "pod: foo-pod in namespace: foo is in pod status phase Pending "},
			wantErr: false,
		},
	}

	// Execute test cases.
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Configure namespace environment variable.
			os.Setenv("TARGET_NAMESPACE", tt.fields.namespace)

			// Create a fake client and options.
			client := fake.NewSimpleClientset(objects...)
			o := Options{client: client}

			// Run the pod status check.
			got, err := o.findPodsNotRunning(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("findPodsNotRunning() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("findPodsNotRunning() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// getTestPods creates namespace and pod fixtures.
func getTestPods() []runtime.Object {
	// Return test namespaces and pods.
	return []runtime.Object{
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "foo",
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo-pod",
				Namespace: "foo",
			},
			Status: v1.PodStatus{
				Phase: v1.PodPending,
			},
		},
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "bar",
			},
		},
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "bar-pod",
				Namespace: "bar",
			},
			Status: v1.PodStatus{
				Phase: v1.PodPending,
			},
		},
	}
}
