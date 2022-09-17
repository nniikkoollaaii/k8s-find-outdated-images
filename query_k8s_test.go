package main

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFilterNamespaces(t *testing.T) {
	filterFlag := "type=workload"
	allNamespaces := corev1.NamespaceList{
		Items: []corev1.Namespace{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "TypeWorkloadNamespace",
					Annotations: map[string]string{
						"type": "workload",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "TypeSystemNamespace",
					Annotations: map[string]string{
						"type": "system",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "NoTypeNamespace",
					Annotations: map[string]string{
						"other": "annotations only",
					},
				},
			},
		},
	}

	result := filterNamespaces(filterFlag, &allNamespaces)

	if len(result) != 1 {
		t.Fatalf("Wrong number of filtered namespaces: %d expected namespaces but got %d", 1, len(result))
	}
}

func TestFilterNamespace_NoFilterFlag(t *testing.T) {
	filterFlag := ""
	allNamespaces := corev1.NamespaceList{
		Items: []corev1.Namespace{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "TypeWorkloadNamespace",
					Annotations: map[string]string{
						"type": "workload",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "TypeSystemNamespace",
					Annotations: map[string]string{
						"type": "system",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "NoTypeNamespace",
					Annotations: map[string]string{
						"other": "annotations only",
					},
				},
			},
		},
	}

	result := filterNamespaces(filterFlag, &allNamespaces)

	if len(result) != 3 {
		t.Fatalf("Wrong number of filtered namespaces: %d expected namespaces but got %d", 1, len(result))
	}
}
func TestFilterNamespace_FilterFlagOnlyKey(t *testing.T) {
	filterFlag := "app"
	allNamespaces := corev1.NamespaceList{
		Items: []corev1.Namespace{
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "App1Namespace",
					Annotations: map[string]string{
						"app": "app1",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "App2Namespace",
					Annotations: map[string]string{
						"app": "app2",
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Name: "NoTypeNamespace",
					Annotations: map[string]string{
						"system": "system1",
					},
				},
			},
		},
	}

	result := filterNamespaces(filterFlag, &allNamespaces)

	if len(result) != 2 {
		t.Fatalf("Wrong number of filtered namespaces: %d expected namespaces but got %d", 1, len(result))
	}
}
