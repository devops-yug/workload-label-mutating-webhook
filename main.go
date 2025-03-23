package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	scheme = runtime.NewScheme()
	codecs = serializer.NewCodecFactory(scheme)
)

func main() {
	http.HandleFunc("/mutateLabels", mutateLabels)
	tlsMode := false
	if tlsEnv, exists := os.LookupEnv("TLS_MODE"); exists && tlsEnv == "true" {
		tlsMode = true
	}
	if tlsMode {
		server := &http.Server{
			Addr: ":443",
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{loadTLSCert()},
			},
		}
		log.Println("****** Server starting in TLS mode on port 443 *******")
		server.ListenAndServeTLS("", "")
	} else {
		log.Println("****** Server starting in HTTP mode on port 8080 *******")
		http.ListenAndServe(":8080", nil)
	}
}

func loadTLSCert() tls.Certificate {
	log.Println("### Setting Environment ###")
	log.Println("### Loading TLS Certs ###")
	cert, err := tls.LoadX509KeyPair("/cert/tls.crt", "/cert/tls.key")
	if err != nil {
		panic(err)
	}
	return cert
}

func mutateLabels(w http.ResponseWriter, r *http.Request) {
	var admissionReview admissionv1.AdmissionReview
	if err := json.NewDecoder(r.Body).Decode(&admissionReview); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pod := &corev1.Pod{}
	if err := json.Unmarshal(admissionReview.Request.Object.Raw, pod); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}

	log.Printf("Patching Pod %s with %s namespace labels", pod.Name, &admissionReview.Request.Namespace)

	config, err := rest.InClusterConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	namespace, err := clientSet.CoreV1().Namespaces().Get(context.TODO(), admissionReview.Request.Namespace, metav1.GetOptions{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	namespaceLabels := namespace.GetLabels()
	// Path to the mounted ConfigMap file
	configFilePath := "/config/ignoreLabels"

	// Load ignore labels
	ignoreLabelList, err := loadIgnoreLabels(configFilePath)
	if err != nil {
		log.Printf("Error loading ignore labels: %v\n", err)
		return
	}

	// Print the populated map (or empty map if file wasn't provided)
	log.Println("Ignore Label List:", ignoreLabelList)

	patches := []map[string]interface{}{}

	if len(pod.Labels) == 0 {
		initialPatch := map[string]interface{}{
			"op":     "add",
			"path":   "metadata/labels",
			"values": map[string]string{},
		}
		patches = append([]map[string]interface{}{initialPatch}, patches...)
	}

	for key, value := range namespaceLabels {
		if ignoreLabelList[key] {
			continue
		}
		if _, exists := pod.Labels[key]; exists {
			patch := map[string]interface{}{
				"op":    "replace",
				"path":  "/metadata/labels/" + key,
				"value": value,
			}
			patches = append(patches, patch)
		} else {
			patch := map[string]interface{}{
				"op":    "add",
				"path":  "metadata/labels/" + key,
				"value": value,
			}
			patches = append(patches, patch)
		}
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	admissionReponse := admissionv1.AdmissionResponse{
		UID:     admissionReview.Request.UID,
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	log.Println("*** Sending response to K8S Api Server")
	admissionReview.Response = &admissionReponse
	if err := json.NewEncoder(w).Encode(admissionReponse); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func loadIgnoreLabels(filePath string) (map[string]bool, error) {
	// Initialize the map
	ignoreLabelList := map[string]bool{}

	// Attempt to open the file
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, return an empty map
			log.Printf("File %s does not exist, ignoring...\n", filePath)
			return ignoreLabelList, nil
		}
		// Other errors while opening the file
		return nil, err
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		label := scanner.Text()
		if label != "" {
			ignoreLabelList[label] = true
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ignoreLabelList, nil
}
