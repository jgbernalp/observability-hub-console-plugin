# Observability UI Hub Plugin for OpenShift Console

This plugin adds a catalog to install Observability UI plugins using the ObservabilityUI operator

## Development

[Node.js](https://nodejs.org/en/) and [npm](https://www.npmjs.com/) are required
to build and run the plugin. To run OpenShift console in a container, either
[Docker](https://www.docker.com) or [podman 3.2.0+](https://podman.io) and
[oc](https://console.redhat.com/openshift/downloads) are required.

### Running locally

In one terminal window, run:

1. `cd web`
1. `npm install`
1. `npm run dev`

In another terminal window, run:

1. `oc login` (requires [oc](https://console.redhat.com/openshift/downloads) and an [OpenShift cluster](https://console.redhat.com/openshift/create))
2. `cd web`
3. `npm run start:console` (requires [Docker](https://www.docker.com) or [podman 3.2.0+](https://podman.io))

This will create an environment file `scripts/env.list` and run the OpenShift console
in a container connected to the cluster you've logged into. The plugin HTTP server
runs on port 9001 with CORS enabled.

Navigate to <http://localhost:9000/observability-ui/catalog> to see the running plugin.

### Running tests

#### Unit tests

```sh
npm run test:unit
```

#### e2e tests

In order to run the e2e tests, you need first to build the plugin in standalone mode

```sh
npm run build:standalone:instrumented
```

and then run the cypress tests

```sh
npm run test:e2e
```

## Deployment on cluster

Install the Observability UI operator on the cluster.

## Build the image

```sh
./scripts/image.sh -t latest
```
