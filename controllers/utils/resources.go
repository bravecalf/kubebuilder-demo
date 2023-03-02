package utils

import (
	"bytes"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	myappv1 "kubebuilder-demo/api/v1"
	"text/template"
)

func parseTemplate(name string, app *myappv1.Foo) []byte {
	tmpl, err := template.ParseFiles(fmt.Sprintf("controllers/templates/%s.yaml", name))
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	err = tmpl.Execute(b, app)
	if err != nil {
		panic(err)
	}
	return b.Bytes()
}

func ConstructDeployment(app *myappv1.Foo) *appsv1.Deployment {
	d := new(appsv1.Deployment)
	err := yaml.Unmarshal(parseTemplate("deployment", app), d)
	if err != nil {
		panic(err)
	}
	return d
}

func ConstructService(app *myappv1.Foo) *corev1.Service {
	s := new(corev1.Service)
	err := yaml.Unmarshal(parseTemplate("service", app), s)
	if err != nil {
		panic(err)
	}
	return s
}

func ConstructIngress(app *myappv1.Foo) *networkingv1.Ingress {
	i := new(networkingv1.Ingress)
	err := yaml.Unmarshal(parseTemplate("ingress", app), i)
	if err != nil {
		panic(err)
	}
	return i
}
