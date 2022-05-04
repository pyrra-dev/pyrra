# Pyrra OpenShift Example

This example walks you through the setup of Pyrra on OpenShift. Please keep 
in mind that doing anything in OpenShift's core namespaces voids support. If 
you want to deploy Pyrra in a production environment, you need to move the 
deployment to a different namespace.

## OpenShift vs Kubernetes
While OpenShift has Kubernetes at its core, there are some significant 
differences. Those get especially obvious when trying to work with 
prometheus-operator: OpenShift comes with its own built-in monitoring stack 
out of the box. Said stack is based on kube-prometheus but has some 
modifications and additions.

### Certificates
The most important changes to a regular installation of Pyrra on Kubernetes are 
that all OpenShift built-in services are tls secured. However, if you didn't 
change anything, those are secured with self-signed certificates.

### Authentication
The Second key difference is the requirement for authentication: Every 
built-in component that's in any way exposed, is secured via an Oauth-proxy 
that is making use of OpenShift's authentication methods. This is generally 
great because you don't have to worry about hardening the exposed part but 
in this case, where we want to hook Pyrra to the built-in monitoring stack, 
that means we need to make sure that Pyrra authenticates accordingly.

## Prerequisites

You need an OpenShift cluster. It generally doesn't matter where it loves 
but for the ease of use we will use a 
[Code Ready Containers](https://crc.dev/crc/) cluster.\
Follow the 
[documentation](https://crc.dev/crc/#installing-codeready-containers_gsg) to 
install CrC on your platform. After tha is done you can run the following 
command to get the command to log in as `kubeadmin`:

```shell
crc console --credentials
```

You can verify that you are logged in, and it's working as expected by 
running the following command which you return `kubeadmin`:

```shell
$ oc whoami
kubeadmin
```

## Deploying Pyrra

Check out this repository and enter the directory like so:

```shell
git clone https://github.com/pyrra-dev/pyrra.git
cd pyrra 
```

As mention in the [Authentication](#authentication) section, Pyrra needs to 
be able to authenticate against OpenShift. In order to do that, you need to 
create a ServiceAccount and get its bearer token. Instead of manually 
getting a serviceAccountToken, we just use
[projected volumes](https://kubernetes.io/docs/concepts/storage/projected-volumes/#serviceaccounttoken)
to automount the token into the pod and be able to use it. This holds the 
big advantage of automatically rotated tokens and also the token never has 
to leave the cluster. In order to deploy Pyrra including the serviceAccount, 
you can run the following commands:

```shell
oc apply -f examples/openshift/deploy/openshift.yaml
oc apply -f examples/openshift/deploy/role.yaml
oc apply -f examples/openshift/deploy/api.yaml
```

If you want to access Pyrra, you can leverage OpenShift's router and create 
a route with the following command:

```shell
oc expose service pyrra-api -n openshift-monitoring
```

This command will return the URL to for you to access Pyrra. In case you 
missed it, you can always find it out later with the following command:

```shell
oc get route pyrra-api -n openshift-monitoring
```

## Deploying SLO Examples

The following command will deploy an example set of SLOs targeting the 
availability of the API-Server using errors as Service Level Indicator:

```shell
oc apply -f examples/openshift/slos/apiserver-request-errors.yaml
```

## Using Pyrra

Maneuver to URL that you get from the following command:

```shell
oc get route pyrra-api -n openshift-monitoring
```
