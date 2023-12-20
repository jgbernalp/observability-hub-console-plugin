package v1

import (
	"context"
	"fmt"
	"net/http"

	"encoding/json"

	validator "github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var log = logrus.WithField("module", "plugins-api")

type PluginResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	IsEnabled   bool   `json:"isEnabled"`
}

type PluginRequest struct {
	Type string `json:"type"`
}

func GetPluginHandler(dynamicClient *dynamic.DynamicClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		pluginName := vars["name"]

		if !validator.IsDNSName(pluginName) {
			log.Error("invalid plugin name")
			http.Error(w, "invalid plugin name", http.StatusBadRequest)
			return
		}

		if len(pluginName) == 0 {
			log.Error("invalid plugin name")
			http.Error(w, "invalid plugin name", http.StatusBadRequest)
			return
		}

		crdGVR := schema.GroupVersionResource{
			Group:    "observability-ui.openshift.io",
			Version:  "v1alpha1",
			Resource: "observabilityuiplugin",
		}

		found, err := dynamicClient.Resource(crdGVR).Namespace("").Get(context.TODO(), pluginName, metav1.GetOptions{})

		if err != nil {
			log.WithError(err).Errorf("observability ui plugin not found: %s", pluginName)
			http.Error(w, "observability ui plugin not found", http.StatusNotFound)
			return
		}

		var cr ObservabilityUIPlugin
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(found.UnstructuredContent(), &cr)

		if err != nil {
			log.WithError(err).Errorf("unable to convert observability ui plugin: %s", pluginName)
			http.Error(w, "unable to convert observability ui plugin", http.StatusInternalServerError)
			return
		}

		response := &PluginResponse{
			Name:        cr.Name,
			DisplayName: cr.Spec.DisplayName,
			Version:     cr.Spec.Version,
			Type:        cr.Spec.Type,
		}

		responseData, err := json.Marshal(response)
		if err != nil {
			log.WithError(err).Error("cannot marshal plugin info")
			http.Error(w, "cannot marshal plugin info", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
	}
}

func availablePlugins() []PluginResponse {
	plugins := []PluginResponse{
		{
			Name:        "logs-observability-ui-plugin",
			DisplayName: "Logs",
			Version:     "dev",
			Type:        "logs",
			IsEnabled:   false,
		},
		{
			Name:        "dashboards-observability-ui-plugin",
			DisplayName: "Dashboards",
			Version:     "dev",
			Type:        "dashboards",
			IsEnabled:   false,
		},
	}

	return plugins
}

func ListPluginHandler(dynamicClient *dynamic.DynamicClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		crdGVR := schema.GroupVersionResource{
			Group:    "observability-ui.openshift.io",
			Version:  "v1alpha1",
			Resource: "observabilityuiplugins",
		}

		found, err := dynamicClient.Resource(crdGVR).Namespace("").List(context.TODO(), metav1.ListOptions{})

		if err != nil {
			log.WithError(err).Errorf("cannot list observability ui plugins")
			http.Error(w, "cannot list ui plugins", http.StatusInternalServerError)
			return
		}

		var crs ObservabilityUIPluginList
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(found.UnstructuredContent(), &crs)

		if err != nil {
			log.WithError(err).Errorf("unable to convert observability ui plugins")
			http.Error(w, "unable to convert observability ui plugins", http.StatusInternalServerError)
			return
		}

		response := make([]PluginResponse, 0)

		for _, availablePlugin := range availablePlugins() {
			for _, plugin := range crs.Items {
				if plugin.Name == availablePlugin.Name {
					availablePlugin.IsEnabled = true
				}
			}
			response = append(response, availablePlugin)
		}

		responseData, err := json.Marshal(response)
		if err != nil {
			log.WithError(err).Error("cannot marshal plugin list info")
			http.Error(w, "cannot marshal plugin list info", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
	}
}

func pluginFromRequest(pluginRequest *PluginRequest) (*ObservabilityUIPlugin, error) {
	var pluginDisplayName string
	var pluginVersion string
	var pluginName string
	var pluginServices []ObservabilityUIPluginService

	switch pluginRequest.Type {
	case "logs":
		pluginDisplayName = "Logs"
		pluginVersion = "dev"
		pluginName = "logs-observability-ui-plugin"
		pluginServices = []ObservabilityUIPluginService{
			{
				Alias:     "backend",
				Name:      "lokistack-dev",
				Namespace: "openshift-logging",
				Port:      8080,
			}}
	case "dashboards":
		pluginDisplayName = "Dashboards"
		pluginVersion = "dev"
		pluginName = "dashboards-observability-ui-plugin"
		pluginServices = nil
	}

	if len(pluginVersion) == 0 {
		return nil, errors.Errorf("invalid plugin type: %s", pluginRequest.Type)
	}

	if len(pluginName) == 0 {
		return nil, errors.Errorf("empty plugin name for plugin type")
	}

	if len(pluginVersion) == 0 {
		return nil, errors.Errorf("no matching version for plugin type")
	}

	obsUIPlugin := &ObservabilityUIPlugin{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ObservabilityUIPlugin",
			APIVersion: "observability-ui.openshift.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pluginName,
			Labels: map[string]string{
				"app.kubernetes.io/name":       pluginName,
				"app.kubernetes.io/instance":   pluginName,
				"app.kubernetes.io/version":    pluginVersion,
				"app.kubernetes.io/part-of":    "observability-ui-operator",
				"app.kubernetes.io/managed-by": "observability-ui-operator",
				"app.kubernetes.io/created-by": "observability-ui-hub",
			},
		},
		Spec: ObservabilityUIPluginSpec{
			DisplayName: pluginDisplayName,
			Version:     pluginVersion,
			Type:        pluginRequest.Type,
			Services:    pluginServices,
		},
	}

	return obsUIPlugin, nil
}

func EnablePluginHandler(dynamicClient *dynamic.DynamicClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var p PluginRequest
		err := dec.Decode(&p)
		if err != nil {
			log.Error("invalid plugin data")
			http.Error(w, "invalid plugin data", http.StatusBadRequest)
			return
		}

		obsUIPlugin, err := pluginFromRequest(&p)

		if err != nil {
			log.WithError(err).Errorf("error creating plugin from request %v", err)
			http.Error(w, fmt.Sprintf("error creating plugin from request: %s", err.Error()), http.StatusNotFound)
			return
		}

		crdGVR := schema.GroupVersionResource{
			Group:    "observability-ui.openshift.io",
			Version:  "v1alpha1",
			Resource: "observabilityuiplugins",
		}

		unstructuredObsUIPluginMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&obsUIPlugin)

		if err != nil {
			log.WithError(err).Errorf("error creating the unstructured observability ui plugin: %s", obsUIPlugin.Name)
			http.Error(w, "error creating the unstructured observability ui plugin", http.StatusNotFound)
			return
		}

		unstructuredObsUIPlugin := &unstructured.Unstructured{Object: unstructuredObsUIPluginMap}

		log.Infof("creating observability ui plugin: %v", unstructuredObsUIPlugin)

		found, err := dynamicClient.Resource(crdGVR).Namespace("").Create(context.Background(), unstructuredObsUIPlugin, metav1.CreateOptions{})

		if err != nil {
			log.WithError(err).Errorf("error creating the observability ui plugin: %s", obsUIPlugin.Name)
			http.Error(w, "error creating the observability ui plugin", http.StatusInternalServerError)
			return
		}

		var cr ObservabilityUIPlugin
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(found.UnstructuredContent(), &cr)

		if err != nil {
			log.WithError(err).Errorf("unable to convert observability ui plugin: %s", obsUIPlugin.Name)
			http.Error(w, "unable to convert observability ui plugin", http.StatusInternalServerError)
			return
		}

		response := &PluginResponse{
			Name:        cr.Name,
			DisplayName: cr.Spec.DisplayName,
			Version:     cr.Spec.Version,
			Type:        cr.Spec.Type,
		}

		responseData, err := json.Marshal(response)
		if err != nil {
			log.WithError(err).Error("cannot marshal plugin info")
			http.Error(w, "cannot marshal plugin info", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(responseData)
	}
}

func DeletePluginHandler(dynamicClient *dynamic.DynamicClient) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		pluginName := vars["name"]

		if !validator.IsDNSName(pluginName) {
			log.Error("invalid plugin name")
			http.Error(w, "invalid plugin name", http.StatusBadRequest)
			return
		}

		if len(pluginName) == 0 {
			log.Error("invalid plugin name")
			http.Error(w, "invalid plugin name", http.StatusBadRequest)
			return
		}

		crdGVR := schema.GroupVersionResource{
			Group:    "observability-ui.openshift.io",
			Version:  "v1alpha1",
			Resource: "observabilityuiplugins",
		}

		// TODO: reuse this code
		_, err := dynamicClient.Resource(crdGVR).Namespace("").Get(context.TODO(), pluginName, metav1.GetOptions{})

		if err != nil {
			log.WithError(err).Errorf("observability ui plugin not found: %s", pluginName)
			http.Error(w, "observability ui plugin not found", http.StatusNotFound)
			return
		}

		err = dynamicClient.Resource(crdGVR).Namespace("").Delete(context.Background(), pluginName, metav1.DeleteOptions{})

		if err != nil {
			log.WithError(err).Errorf("error deleting the observability ui plugin: %s", pluginName)
			http.Error(w, "error deleting the observability ui plugin", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}
}
