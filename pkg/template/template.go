package template

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/lastbackend/registry/pkg/registry/context"
	"io/ioutil"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/apis/extensions"
	"os"
	"path/filepath"
	"strings"
)

var storage map[string]map[string]*Template = make(map[string]map[string]*Template)

type Template struct {
	Namespaces             []api.Namespace             `json:"namespaces,omitempty"`
	PersistentVolumes      []api.PersistentVolume      `json:"persistent_volumes,omitempty"`
	PersistentVolumeClaims []api.PersistentVolumeClaim `json:"persistent_volume_claims,omitempty"`
	ServiceAccounts        []api.ServiceAccount        `json:"service_accounts,omitempty"`
	Services               []api.Service               `json:"services,omitempty"`
	ReplicationControllers []api.ReplicationController `json:"replication_controllers,omitempty"`
	Pods                   []api.Pod                   `json:"pods,omitempty"`
	Deployments            []extensions.Deployment     `json:"deployments,omitempty"`
}

func (t *Template) ToJson() ([]byte, error) {
	return json.Marshal(t)
}

func Load(path string) {

	var (
		template string
		version  string
		skip     = true
		ctx      = context.Get()
	)

	ctx.Log.Info("Load templates")

	// walk all files in directory
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {

		if skip {
			skip = false
			return nil
		}

		if info.IsDir() && version != "" {
			template = ""
			version = ""
		}

		if info.IsDir() && template == "" {
			template = info.Name()
			storage[template] = make(map[string]*Template)

		} else if info.IsDir() && version == "" {
			version = info.Name()

			storage[template][version] = new(Template)

		} else {

			yml, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Println(err)
			}

			var parts = strings.Split(string(yml), "---")

			for index := range parts {

				var a = new(struct {
					Kind string `json:"kind"`
				})

				err := yaml.Unmarshal([]byte(parts[index]), a)
				if err != nil {
					fmt.Println(err)
				}

				buf, err := yaml.YAMLToJSON([]byte(parts[index]))
				if err != nil {
					fmt.Println(err)
				}

				switch a.Kind {
				case "Namespace":
					var namespace = new(api.Namespace)

					err := json.Unmarshal(buf, namespace)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].Namespaces = append(storage[template][version].Namespaces, *namespace)
				case "PersistentVolume":
					var persistentVolume = new(api.PersistentVolume)

					err := json.Unmarshal(buf, persistentVolume)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].PersistentVolumes = append(storage[template][version].PersistentVolumes, *persistentVolume)
				case "PersistentVolumeClaim":
					var persistentVolumeClaim = new(api.PersistentVolumeClaim)

					err := json.Unmarshal(buf, persistentVolumeClaim)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].PersistentVolumeClaims = append(storage[template][version].PersistentVolumeClaims, *persistentVolumeClaim)
				case "ServiceAccount":
					var serviceAccount = new(api.ServiceAccount)

					err := json.Unmarshal(buf, serviceAccount)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].ServiceAccounts = append(storage[template][version].ServiceAccounts, *serviceAccount)
				case "Service":
					var service = new(api.Service)

					err := json.Unmarshal(buf, service)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].Services = append(storage[template][version].Services, *service)
				case "Deployment":
					var deployment = new(extensions.Deployment)

					err := json.Unmarshal(buf, deployment)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].Deployments = append(storage[template][version].Deployments, *deployment)
				case "ReplicationController":
					var replicationController = new(api.ReplicationController)

					err := json.Unmarshal(buf, replicationController)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].ReplicationControllers = append(storage[template][version].ReplicationControllers, *replicationController)
				case "Pod":
					var pod = new(api.Pod)

					err := json.Unmarshal(buf, pod)
					if err != nil {
						fmt.Println(err)
					}

					storage[template][version].Pods = append(storage[template][version].Pods, *pod)
				default:
					return nil
				}
			}
		}

		return nil
	})

	buf, err := json.Marshal(storage)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("template: >> ", string(buf))

	return
}

func Get(name, version string) *Template {

	if _, ok := storage[name]; !ok {
		return nil
	}

	if _, ok := storage[name][version]; !ok {
		return nil
	}

	return storage[name][version]
}

func List() map[string][]string {

	var lists = make(map[string][]string)

	for name, versions := range storage {
		for version := range versions {
			lists[name] = append(lists[name], version)
		}
	}

	return lists
}
