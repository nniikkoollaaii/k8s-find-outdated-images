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

func TestGetNamespaceData(t *testing.T) {
	emailFlag := "email"
	namespaces := make(map[string]corev1.Namespace)
	namespaces["CorrectAnnotationNamespace"] = corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "CorrectAnnotationNamespace",
			Annotations: map[string]string{
				"email": "email@domain.de",
			},
		},
	}
	namespaces["WrongAnnotationNamespace"] = corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "WrongAnnotationNamespace",
			Annotations: map[string]string{
				"iwas": "email2@domain.de",
			},
		},
	}
	namespaces["NoAnnotationNamespace"] = corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "NoAnnotationNamespace",
		},
	}

	result := make(map[string]NotificationData)

	getNamespaceData(emailFlag, namespaces, &result)

	if result["CorrectAnnotationNamespace"].Email != "email@domain.de" {
		t.Fatalf("Expected email set for namespace %s", "CorrectAnnotationNamespace")
	}
	if result["WrongAnnotationNamespace"].Email != "" {
		t.Fatalf("Expected no email set for namespace %s", "WrongAnnotationNamespace")
	}
	if result["NoAnnotationNamespace"].Email != "" {
		t.Fatalf("Expected no email set for namespace %s", "NoAnnotationNamespace")
	}
}

func TestImageDataFromK8sAPI(t *testing.T) {
	pods := corev1.PodList{
		Items: []corev1.Pod{
			{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "ns1",
					Name:      "app1",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "my.domain.com/initcontainer1:v1",
						},
						{
							Image: "my.domain.com/initcontainer2:v1",
						},
					},
					Containers: []corev1.Container{
						{
							Image: "my.domain.com/container1:v1",
						},
						{
							Image: "my.domain.com/container2:v1",
						},
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "ns2",
					Name:      "app2",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "my.domain.com/initcontainer2:v1",
						},
						{
							Image: "my.domain.com/initcontainer3:v1",
						},
					},
					Containers: []corev1.Container{
						{
							Image: "my.domain.com/container2:v1",
						},
						{
							Image: "my.domain.com/container3:v1",
						},
					},
				},
			},
			{
				ObjectMeta: v1.ObjectMeta{
					Namespace: "ns3",
					Name:      "app3",
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{
						{
							Image: "my.domain.com/initcontainer3:v1",
						},
						{
							Image: "my.domain.com/initcontainer4:v1",
						},
					},
					Containers: []corev1.Container{
						{
							Image: "my.domain.com/container3:v1",
						},
						{
							Image: "my.domain.com/container4:v1",
						},
					},
				},
			},
		},
	}

	allImages := make(map[string]ImageData)

	addImageData(&allImages, &pods)

	if len(allImages) != 8 {
		t.Fatalf("Wrong number of detected images: %d expected but got %d", 8, len(allImages))
	}
	if len(allImages["my.domain.com/container1:v1"].Findings) != 1 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/container1:v1", 1, len(allImages["my.domain.com/container1:v1"].Findings))
	}
	if len(allImages["my.domain.com/container2:v1"].Findings) != 2 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/container2:v1", 2, len(allImages["my.domain.com/container2:v1"].Findings))
	}
	if len(allImages["my.domain.com/container3:v1"].Findings) != 2 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/container3:v1", 2, len(allImages["my.domain.com/container3:v1"].Findings))
	}
	if len(allImages["my.domain.com/container4:v1"].Findings) != 1 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/container4:v1", 1, len(allImages["my.domain.com/container4:v1"].Findings))
	}
	if len(allImages["my.domain.com/initcontainer1:v1"].Findings) != 1 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/initcontainer1:v1", 1, len(allImages["my.domain.com/initcontainer1:v1"].Findings))
	}
	if len(allImages["my.domain.com/initcontainer2:v1"].Findings) != 2 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/initcontainer2:v1", 2, len(allImages["my.domain.com/initcontainer2:v1"].Findings))
	}
	if len(allImages["my.domain.com/initcontainer3:v1"].Findings) != 2 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/initcontainer3:v1", 2, len(allImages["my.domain.com/initcontainer3:v1"].Findings))
	}
	if len(allImages["my.domain.com/initcontainer4:v1"].Findings) != 1 {
		t.Fatalf("Wrong number of findings for image %s: %d expected but got %d", "my.domain.com/initcontainer4:v1", 1, len(allImages["my.domain.com/initcontainer4:v1"].Findings))
	}
}
