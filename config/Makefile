# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#
# Allows for resources to be loaded from outside the root location of
# the kustomize config file. Ensures that resource don't need to be
# copied around the file system.
#
# See https://kubectl.docs.kubernetes.io/faq/kustomize
#
KOPTIONS := --load-restrictor LoadRestrictionsNone

#
# Include the main camel-k Makefile containing
# basic common recipes like kustomize and vars
# like VERSION
#
include script/Makefile

#
# Vars that can be overridden by external env vars
#
DEBUG ?= false
TAG ?= $(VERSION)
NAMESPACE ?= camel-k

# Image pull policy: [IfNotPresent|Always]
IMAGE_PULL_POLICY ?= IfNotPresent
# Global: [true|false]
# - On setup: will promote roles and bindings to cluster-level
# - On operator: set namespace to all using WATCH_NAMESPACE env var
GLOBAL ?= false
# Monitoring: [true|false]
# - On operator: will add the prometheus resources to install
MONITORING ?= false
# Monitoring Port: integer
MONITORING_PORT ?= 8080
# Health Port: integer
HEALTH_PORT ?= 8081

RBAC_KUBE := rbac/kubernetes
RBAC_OS := rbac/openshift
RBAC_CLUSTER := cluster
RBAC_GLOBAL := global
INFRASTRUCTURE := infrastructure
YAML := yaml

OP_PATCHES := ./operator/patches
PLATFORM_PATCHES := ./platform/patches
PLACEHOLDER := placeholder
IMAGE_PULL_POLICY_PATCH := image-pull-policy-always-patch
WATCH_NAMESPACE_PATCH := watch-namespace-global-patch
PORTS_PATCH := ports-patch
INT_PLATFORM_PATCH := integration-platform-patch

#
# Macro for editing kustomization to define
# the image reference
#
# Parameter: directory of the kustomization.yaml
#
define set-kustomize-image
	$(if $(filter $(DEFAULT_IMAGE),$(IMAGE_NAME):$(TAG)),,\
		@cd $(1) && $(KUSTOMIZE) edit set image $(DEFAULT_IMAGE)=$(IMAGE_NAME):$(TAG))
endef

#
# Macro for editing kustomization to define
# the namespace
#
# Parameter: directory of the kustomization.yaml
#
define set-kustomize-namespace
	@cd $(1) && $(KUSTOMIZE) edit set namespace $(NAMESPACE)
endef

#
# Macro for swapping a resource directory to an alternative directory
#
# Parameters:
# * directory of the kustomization.yaml
# * source directory
# * target directory
#
define swap-rbac
	@cd $(1) && \
		$(KUSTOMIZE) edit remove resource ../$(2) && \
		$(KUSTOMIZE) edit add resource ../$(3)
endef

#
# Add or remove a patch on a kustomization.yaml
# targetting the deployment resource
#
# Parameters:
# * directory of the kustomization.yaml
# * [add, remove]
# * path of patch
#
define add-remove-deployment-patch
	@cd $(1) && \
		$(KUSTOMIZE) edit $(2) patch --path patches/$(3) --kind Deployment --name camel-k-operator
endef

#
# Macro for adding / removing the prometheus resources for monitoring
#
define add-remove-operator-monitoring
	cd $(1) && \
		$(KUSTOMIZE) edit $(2) resource ../prometheus
endef

.PHONY: have-platform check_admin setup-cluster .setup-kubernetes .setup-openshift setup

#
# Determine the platform of the cluster based on
# either the use of querying through a go-client
# or using an installed client, ie. oc or kubectl
#
find-platform:
ifndef PLATFORM
PLATFORM=$(shell script/check_platform.sh)
endif

#
# Checks if the cluster platform has been defined correctly either by the user
# or by the platform_check script.
#
have-platform: find-platform
ifeq ($(PLATFORM),openshift)
	@echo Platform identified as 'openshift'
else ifeq ($(PLATFORM),kubernetes)
	@echo Platform identified as 'kubernetes'
else
	@echo "****"
	@echo "**** ERROR: Cannot continue as cluster platform cannot be identified ****"
	@echo "****"
	@exit 1
endif

#
# Checks if the cluster user has the necessary privileges to be a cluster-admin
# In this case if the user can list the CRDs then probably a cluster-admin
#
check-admin: kubectl
	@output=$$(kubectl get crd 2>&1) || (echo "****" && echo "**** ERROR: Cannot continue as user is not a Cluster-Admin ****" && echo "****"; exit 1)

#
# Setup the cluster installation by installing crds and cluster roles.
#
# Will either call setup-cluster-openshift (then setup-cluster-kubernetes) or
# setup-cluster-kubernetes depending on the identity of the cluster
#
# PARAMETERS:
#   NAMESPACE: Sets the namespace for the resources
#   PLATFORM:  Override the discovered platform, if required
#   DEBUG:     true - Prints the resources to be applied instead of applying them
#
setup-cluster: have-platform kustomize kubectl check-admin
# Set the namespace in the setup-cluster kustomization yaml
	@$(call set-kustomize-namespace,$@)
#
#
ifeq ($(PLATFORM), openshift)
	@cd $@ && $(KUSTOMIZE) edit add resource ../$(RBAC_OS)-$(RBAC_CLUSTER)
endif
#
# Build the resources
# Post-process ClusterRoleBindings to fix the namespace in the refs (not yet handled by kustomize)
# Either apply to the cluster or output to CLI
#
ifeq ($(DEBUG), false)
	@$(KUSTOMIZE) build $(KOPTIONS) $@ | \
		sed 's/$(PLACEHOLDER)/$(NAMESPACE)/' | \
		kubectl apply -f -
else
	@$(KUSTOMIZE) build $(KOPTIONS) $@ | \
		sed 's/$(PLACEHOLDER)/$(NAMESPACE)/'
endif

#
# Modify the rbac/kubernetes resource definition depending on GLOBAL
#
.setup-kubernetes:
# Resets back to default by removing any reference to rbac/openshift
	@cd setup && \
		$(KUSTOMIZE) edit remove resource ../$(RBAC_OS) && \
		$(KUSTOMIZE) edit remove resource ../$(RBAC_OS)-$(RBAC_GLOBAL)
ifeq ($(GLOBAL),true)
	@echo "****"
	@echo "**** GLOBAL requires a user with cluster-admin privileges ****"
	@echo "****"
	@$(MAKE) -s check-admin
	@$(call swap-rbac,setup,"$(RBAC_KUBE)","$(RBAC_KUBE)-$(RBAC_GLOBAL)")
else
	@$(call swap-rbac,setup,"$(RBAC_KUBE)-$(RBAC_GLOBAL)","$(RBAC_KUBE)")
endif

#
# Add in the rbac role designated in infrastructure
# Modify the rbac/openshift resource definition depending on GLOBAL
#
.setup-openshift: .setup-kubernetes
	@cd $(INFRASTRUCTURE) && $(KUSTOMIZE) edit add resource $(RBAC_OS)
ifeq ($(GLOBAL),true)
	@echo "****"
	@echo "**** GLOBAL requires a user with cluster-admin privileges ****"
	@echo "****"
	@$(MAKE) -s check-admin
	@$(call swap-rbac,setup,"$(RBAC_OS)","$(RBAC_OS)-$(RBAC_GLOBAL)")
else
	@$(call swap-rbac,setup,"$(RBAC_OS)-$(RBAC_GLOBAL)","$(RBAC_OS)")
endif

#
# Setup the installation by installing roles and granting
# privileges for the installing operator. Although a cluster-admin
# is not required a suitably privieged user is in order to create all
# the required namespace roles.
#
# PARAMETERS:
#   NAMESPACE: Sets the namespace for the resources
#   GLOBAL:    Converts all roles & bindings to cluster-level (requires cluster-admin privileges)
#   PLATFORM:  Override the discovered platform, if required
#   DEBUG:     true - Prints the resources to be applied instead of applying them
#
setup: have-platform kustomize kubectl
# Cannot be a dependency as PLATFORM could contain 'ERROR: '
	@$(MAKE) .setup-$(PLATFORM)
# Set the namespace in the setup kustomization yaml
	@$(call set-kustomize-namespace,$@)
#
# Build the resources
# Post-process RoleBindings to fix the namespace in the refs (not yet handled by kustomize)
# Either apply to the cluster or output to CLI
#
ifeq ($(DEBUG), false)
	@$(KUSTOMIZE) build $(KOPTIONS) $@ | \
		sed 's/$(PLACEHOLDER)/$(NAMESPACE)/' | \
		kubectl apply -f -
else
	@$(KUSTOMIZE) build $(KOPTIONS) $@ | \
		sed 's/$(PLACEHOLDER)/$(NAMESPACE)/'
endif

.PHONY: operator .operator-port-patch .operator-can-monitor

#
# Customizes the port patch
#
.operator-port-patch:
	@sed -i 's/--monitoring-port=.*/--monitoring-port=$(MONITORING_PORT)/' $(OP_PATCHES)/$(PORTS_PATCH).$(YAML)
	@sed -i '/path:.*\/containerPort/,/- op/{s/value: .*/value: $(MONITORING_PORT)/}' $(OP_PATCHES)/$(PORTS_PATCH).$(YAML)
	@sed -i 's/--health-port=.*/--health-port=$(HEALTH_PORT)/' $(OP_PATCHES)/$(PORTS_PATCH).$(YAML)
	@sed -i '/path:.*\/httpGet\/port/,/- op/{s/value: .*/value: $(HEALTH_PORT)/}' $(OP_PATCHES)/$(PORTS_PATCH).$(YAML)

.operator-can-monitor: kubectl
	@output=$$(kubectl get crd prometheusrules.monitoring.coreos.com 2>&1) || (echo "****" && echo "**** ERROR: Montoring not available as CRDs not installed in cluster ****" && echo "****"; exit 1)

#
# Install the operator deployment and related resources
#
# PARAMETERS:
#   MONITORING:        Adds the prometheus monitoring resources
#   IMAGE_NAME:        Set a custom operator image name
#   TAG:               Set a custom operator image tag
#   NAMESPACE:         Set the namespace to install the operator into
#   GLOBAL:            Sets the operator to watch all namespaces for custom resources
#   PLATFORM:          Override the discovered platform, if required
#   IMAGE_PULL_POLICY: Sets whether to always pull the operator image
#   MONITORING_PORT:   Set a custom monitoring port
#   HEALTH_PORT:       Set a custom health port
#   DEBUG:             Prints the resources to be applied instead of applying them
#
operator: have-platform kustomize kubectl .operator-port-patch
ifeq ($(MONITORING), true)
	@echo "****"
	@echo "**** MONITORING requires a user with cluster-admin privileges ****"
	@echo "****"
	@$(MAKE) -s check-admin
	@$(MAKE) -s .operator-can-monitor
	@$(call add-remove-operator-monitoring,$@,add)
else
	@$(call add-remove-operator-monitoring,$@,remove)
endif
# Set the namespace in the setup kustomization yaml
	@$(call set-kustomize-namespace,$@)
# Set the image reference of the kustomization
	@$(call set-kustomize-image,$@)
# Set the WATCH NAMESPACE env var depending on GLOBAL var
ifeq ($(GLOBAL), true)
	@$(call add-remove-deployment-patch,$@,add,$(WATCH_NAMESPACE_PATCH).$(YAML))
else
	@$(call add-remove-deployment-patch,$@,remove,$(WATCH_NAMESPACE_PATCH).$(YAML))
endif
# Set the IMAGE_PULL_POLICY config depending on var
ifeq ($(IMAGE_PULL_POLICY),Always)
	@$(call add-remove-deployment-patch,$@,add,$(IMAGE_PULL_POLICY_PATCH).$(YAML))
else
	@$(call add-remove-deployment-patch,$@,remove,$(IMAGE_PULL_POLICY_PATCH).$(YAML))
endif
# Set the PORTS depending on vars
ifneq ($(MONITORING_PORT), 8080)
	@$(call add-remove-deployment-patch,$@,add,$(PORTS_PATCH).$(YAML))
else ifneq ($(HEALTH_PORT), 8081)
	@$(call add-remove-deployment-patch,$@,add,$(PORTS_PATCH).$(YAML))
endif
ifeq ($(DEBUG), false)
	@$(KUSTOMIZE) build $(KOPTIONS) $@ | kubectl apply -f -
else
	@$(KUSTOMIZE) build $(KOPTIONS) $@
endif

.PHONY: platform .platform-openshift-patch .platform-kubernetes-patch

#
# Customizes the samples patches for kubernetes
#
.platform-kubernetes-patch:
	@sed -i 's/.*profile:.*/  profile: Kubernetes/' $(PLATFORM_PATCHES)/$(INT_PLATFORM_PATCH).$(YAML)

#
# Customizes the samples patches for openshift
#
.platform-openshift-patch:
	@sed -i 's/.*profile:.*/  profile: Openshift/' $(PLATFORM_PATCHES)/$(INT_PLATFORM_PATCH).$(YAML)

#
# Install the integration platform
#
# PARAMETERS:
#   NAMESPACE: Set the namespace to install the operator into
#   PLATFORM:  Override the discovered platform, if required
#   DEBUG:     true - Prints the resources to be applied instead of applying them
#
platform: have-platform kustomize kubectl
# Cannot be a dependency as PLATFORM could contain 'ERROR: '
	@$(MAKE) .platform-$(PLATFORM)-patch
# Set the namespace in the setup kustomization yaml
	@$(call set-kustomize-namespace,$@)
ifeq ($(DEBUG), false)
	@$(KUSTOMIZE) build $(KOPTIONS) $@ | kubectl apply -f -
else
	@$(KUSTOMIZE) build $(KOPTIONS) $@
endif

.PHONY: example

#
# Installs the operator deployment and the example integration
#
# PARAMETERS:
#   NAMESPACE: Set the namespace to install the operator into
#   PLATFORM:  Override the discovered platform, if required
#   DEBUG:     true - Prints the resources to be applied instead of applying them
#
example: kubectl operator
# Set the namespace in the setup kustomization yaml
	@$(call set-kustomize-namespace,$@)
ifeq ($(DEBUG), false)
	@$(KUSTOMIZE) build $(KOPTIONS) $@ | kubectl apply -f -
else
	@$(KUSTOMIZE) build $(KOPTIONS) $@
endif
