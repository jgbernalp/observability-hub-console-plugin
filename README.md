# Observability Plugins Hub for OpenShift Console

This plugin adds the observability UI plugins hub into the OpenShift console. It requires OpenShift 4.12.

## Development

[Node.js](https://nodejs.org/en/) and [npm](https://www.npmjs.com/) are required
to build and run the plugin. To run OpenShift console in a container, either
[Docker](https://www.docker.com) or [podman 3.2.0+](https://podman.io) and
[oc](https://console.redhat.com/openshift/downloads) are required.

### Running locally

Make sure you have loki running on `http://localhost:3100`

1. Install the dependencies running `make install`
2. Start the backend `make start-backend`
3. In a different terminal start the frontend `make start-frontend`
4. In a different terminal start the console
   a. `oc login` (requires [oc](https://console.redhat.com/openshift/downloads) and an [OpenShift cluster](https://console.redhat.com/openshift/create))
   b. `make start-console` (requires [Docker](https://www.docker.com) or [podman 3.2.0+](https://podman.io))

This will create an environment file `web/scripts/env.list` and run the OpenShift console
in a container connected to the cluster you've logged into. The plugin backend server
runs on port 9002 with CORS enabled.

Navigate to <http://localhost:9000/observability-ui/catalog> to see the catalog of plugins

### Running tests

#### Unit tests

```sh
make test-unit
```

#### e2e tests

```sh
make test-frontend
```

this will build the frontend in standalone mode and run the cypress tests

## Deployment on cluster

You can deploy the plugin to a cluster by instantiating the provided
[Plugin Resources](observability-ui-hub-resources.yml). It will use the latest plugin
docker image and run a light-weight go HTTP server to serve the plugin's assets.

```sh
oc create -f observability-ui-hub-resources.yml
```

Once deployed, patch the [Console operator](https://github.com/openshift/console-operator)
config to enable the plugin.

```sh
oc patch consoles.operator.openshift.io cluster \
  --patch '{ "spec": { "plugins": ["observability-ui-hub"] } }' --type=merge
```

## Plugin configuration

The plugin can be configured by mounting a ConfigMap in the deployment and passing the `-plugin-config-path` flag with the file path, for example:

ConfigMap with plugin configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: observability-ui-hub-config
  namespace: openshift-observability-ui
  labels:
    app: observability-ui-hub
    app.kubernetes.io/part-of: observability-ui-hub
data:
  config.yaml: |-
    timeout: '60s'
```

Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: observability-ui-hub
  namespace: openshift-observability-ui
  labels:
    app: observability-ui-hub
    app.kubernetes.io/component: observability-ui-hub
    app.kubernetes.io/instance: observability-ui-hub
    app.kubernetes.io/part-of: observability-ui-hub
    app.openshift.io/runtime-namespace: openshift-observability-ui
spec:
  replicas: 1
  selector:
    matchLabels:
      app: observability-ui-hub
  template:
    metadata:
      labels:
        app: observability-ui-hub
    spec:
      containers:
        - name: observability-ui-hub
          image: "quay.io/gbernal/observability-ui-hub:latest"
          args:
            - "-plugin-config-path"
            - "/etc/plugin/config.yaml"
            ...

          volumeMounts:
            - name: plugin-config
              readOnly: true
              mountPath: /etc/plugin/config.yaml
              subPath: config.yaml
            ...

      volumes:
        - name: plugin-conf
          configMap:
            name: observability-ui-hub-config
            defaultMode: 420
        ...

      ...

```

# Configuration values

| Field   | Description                        | Default | Unit                                         |
| :------ | :--------------------------------- | :------ | :------------------------------------------- |
| timeout | fetch timeout when requesting logs | `30s`   | [duration](https://pkg.go.dev/time#Duration) |

## Build a testint the image

```sh
make build-image
```
