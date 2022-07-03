package main

import (
	"context"
	"encoding/json"

	// "flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Config struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	UpdatedAt   string `json:"updatedAt"`
	ConfigURL   string `json:"configURL"`
	User        string `json:"user"`
}

type Response struct {
	Status string `json:"status"`
}

type Deployment struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
	Condition string `json:"condition"`
}

type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Ports     string `json:"ports"`
}

var clientSet *kubernetes.Clientset = nil

func downloadConfig(user string, configURL string, configName string, uploadPath string) error {
	fmt.Println("Downloading config from: ", configURL)
	fmt.Print("\n")

	// Get the data
	resp, err := http.Get(configURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	err = os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		return err
	}
	out, err := os.Create(uploadPath + "/" + configName)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func testCluster(configAbsLoc string) error {
	fmt.Println("Testing cluster: " + configAbsLoc)
	fmt.Print("\n")

	kubeconfig, err := ioutil.ReadFile(configAbsLoc)
	if err != nil {
		return err
	}
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		return err
	}
	clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	pods, err := clientSet.CoreV1().Pods("lucifergene").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Println("Pods: ")
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}

	return err
}

func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

	w.WriteHeader(http.StatusOK)
	r.ParseForm()

	var config Config
	config.Name = r.FormValue("name")
	config.DisplayName = r.FormValue("displayName")
	config.UpdatedAt = r.FormValue("updatedAt")
	config.ConfigURL = r.FormValue("configURL")
	config.User = r.FormValue("user")

	var response Response
	response.Status = "OK"

	// err :=  json.NewDecoder(r.Body).Decode(&config)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(config)

	uploadPath := "uploads/" + config.User

	err := downloadConfig(config.User, config.ConfigURL, config.Name, uploadPath)
	if err != nil {
		panic(err)
	}
	fmt.Println("Downloaded: " + config.ConfigURL)

	uploadedConfigLocation, err := filepath.Abs(uploadPath + "/" + config.Name)
	if err != nil {
		panic(err)
	}
	testErr := testCluster(uploadedConfigLocation)

	if testErr != nil {
		response.Status = testErr.Error()
	}
	json.NewEncoder(w).Encode(response)
}

func DisconnectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

	w.WriteHeader(http.StatusOK)

	if clientSet != nil {
		clientSet = nil
	}

	var response Response
	response.Status = "OK"

	json.NewEncoder(w).Encode(response)

}

//DEPLOYMENTS
func getDeployments() ([]Deployment, error) {
	fmt.Println("Getting deployments :")
	fmt.Print("\n")

	var deploymentArray []Deployment = nil

	if clientSet == nil {
		return deploymentArray, fmt.Errorf("cluster not connected")
	}

	deployments, err := clientSet.AppsV1().Deployments("lucifergene").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return deploymentArray, err
	}
	for _, deployment := range deployments.Items {
		var deploymentObject Deployment
		deploymentObject.Name = deployment.Name
		deploymentObject.Namespace = deployment.Namespace
		deploymentObject.Replicas = deployment.Status.Replicas
		deploymentObject.Condition = string(deployment.Status.Conditions[0].Type)
		deploymentArray = append(deploymentArray, deploymentObject)
	}

	fmt.Println("Deployments: ")
	for _, deployment := range deployments.Items {
		fmt.Println(deployment.Name, deployment.Namespace, string(deployment.Status.Conditions[0].Type), deployment.Status.Replicas)
	}
	return deploymentArray, err
}

func DeploymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

	w.WriteHeader(http.StatusOK)

	deploymentArray, err := getDeployments()
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
	json.NewEncoder(w).Encode(deploymentArray)
}

//SERVICES
func getServices() ([]Service, error) {
	fmt.Println("Getting services :")
	fmt.Print("\n")

	var serviceArray []Service = nil

	if clientSet == nil {
		return serviceArray, fmt.Errorf("cluster not connected")
	}

	services, err := clientSet.CoreV1().Services("lucifergene").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return serviceArray, err
	}
	for _, service := range services.Items {
		var serviceObject Service
		portString := fmt.Sprintf("%d:%s/%s", service.Spec.Ports[0].Port, service.Spec.Ports[0].TargetPort.String(), string(service.Spec.Ports[0].Protocol))
		serviceObject.Name = service.Name
		serviceObject.Namespace = service.Namespace
		serviceObject.Type = string(service.Spec.Type)
		serviceObject.Ports = portString
		serviceArray = append(serviceArray, serviceObject)
	}

	fmt.Println("Services: ")
	for _, service := range services.Items {
		fmt.Println(service.Name, service.Namespace, string(service.Spec.Type), service.Spec.Ports[0].Port, service.Spec.Ports[0].TargetPort.String(), string(service.Spec.Ports[0].Protocol))
	}
	return serviceArray, err
}

func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")

	w.WriteHeader(http.StatusOK)

	serviceArray, err := getServices()
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}
	json.NewEncoder(w).Encode(serviceArray)
}
