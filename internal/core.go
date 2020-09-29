/*
 * Copyright 2020 Skulup Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package internal

import (
	"github.com/skulup/operator-pkg/k8s"
	"github.com/skulup/operator-pkg/k8s/configmap"
	"github.com/skulup/operator-pkg/types"
	"github.com/skulup/operator-pkg/util"
	v12 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

// OperatorName defines the name of the operator
const OperatorName = "pulsar-operator"

// Domain defines the domain of the operator
const Domain = "pulsar.skulup.com"

const (
	// The broker app
	Broker = "pulsar-broker"
	// The proxy app
	Proxy = "pulsar-proxy"
	// The manager app
	Manager = "pulsar-manager"
)

// The label indicating the cluster a broker belongs to
const LabelCluster = "pulsar.skulup.com/cluster"
const annPrometheusScrape = "prometheus.io/scrape"
const annPrometheusPort = "prometheus.io/port"

var (
	// ServicePort The broker non-TLS port
	ServicePort = util.Int32Or("BROKER_SERVICE_PORT", 6650)
	// ServicePortTLS The broker TLS port
	ServicePortTLS = util.Int32Or("BROKER_SERVICE_PORT_TLS", 6651)
	// WebServicePort The broker non-TLS web port
	WebServicePort = util.Int32Or("BROKER_WEB_SERVICE_PORT", 8080)
	// WebServicePortTLS The broker TLS web port
	WebServicePortTLS = util.Int32Or("BROKER_WEB_SERVICE_PORT_TLS", 8443)
	// ManagerBackendPort The manager API port
	ManagerBackendPort = util.Int32Or("MANAGER_BACKEND_SERVICE_PORT", 7750)
	// ManagerFrontendPort The manager UI port
	ManagerFrontendPort = util.Int32Or("MANAGER_FRONTEND_SERVICE_PORT", 9527)
)

// IsApplyConfigFromEnvScriptFaulty returns false for image tag >= 2.6.0 and true otherwise.
// Pulsar below version 2.6.0 has a buggy `apply-config-from-env.py` - it fails to update
// existing broker.conf configuration parameters by using the `PULSAR_PREFIX_` prefix environment variables.
// https://github.com/apache/pulsar/blob/v2.5.2/docker/pulsar/scripts/apply-config-from-env.py
func IsApplyConfigFromEnvScriptFaulty(img types.Image) bool {
	tag := img.Tag
	if tag == "latest" {
		return false
	}
	if nTag, err := strconv.Atoi(strings.ReplaceAll(tag, ".", "")); err == nil {
		return nTag < 260
	}
	return true
}

// NewPulsarConfigMap returns a new ConfigMap
func NewPulsarConfigMap(img types.Image, namespace, name string) *v12.ConfigMap {
	data := map[string]string{}
	if IsApplyConfigFromEnvScriptFaulty(img) {
		data["apply-config-from-env.py"] = applyConfigFromEnvPythonScript
	}
	return configmap.New(namespace, name, data)
}

// NewPulsarConfigMap returns the default pulsar service ports
func CreateServicePorts() []v12.ServicePort {
	return []v12.ServicePort{
		{
			Name:     "pulsar-tcp",
			Protocol: v12.ProtocolTCP,
			Port:     ServicePort,
		},
		{
			Name:     "pulsar-tls",
			Protocol: v12.ProtocolTCP,
			Port:     ServicePortTLS,
		},
		{
			Name:     "pulsar-http",
			Protocol: v12.ProtocolTCP,
			Port:     WebServicePort,
		},
		{
			Name:     "pulsar-https",
			Protocol: v12.ProtocolTCP,
			Port:     WebServicePortTLS,
		},
	}
}

// GenerateLabels create labels for the different pulsar components
func GenerateLabels(appName string, more map[string]string) map[string]string {
	labels := map[string]string{
		k8s.LabelAppManagedBy: OperatorName,
	}
	if appName != "" {
		labels[k8s.LabelAppName] = appName
	}
	for k, v := range more {
		labels[k] = v
	}
	return labels
}

// GenerateAnnotations create labels for the different pulsar components
func GenerateAnnotations(source map[string]string) map[string]string {
	annotations := map[string]string{}
	for k, v := range source {
		annotations[k] = v
	}
	if annotations[annPrometheusScrape] == "" {
		annotations[annPrometheusScrape] = "true"
	}
	if annotations[annPrometheusPort] == "" {
		annotations[annPrometheusPort] = "8080"
	}
	return annotations
}

const applyConfigFromEnvPythonScript = "#!/usr/bin/env python\n#\n# Licensed to the Apache Software Foundation (ASF) under one\n# or more contributor license agreements.  See the NOTICE file\n# distributed with this work for additional information\n# regarding copyright ownership.  The ASF licenses this file\n# to you under the Apache License, Version 2.0 (the\n# \"License\"); you may not use this file except in compliance\n# with the License.  You may obtain a copy of the License at\n#\n#   http://www.apache.org/licenses/LICENSE-2.0\n#\n# Unless required by applicable law or agreed to in writing,\n# software distributed under the License is distributed on an\n# \"AS IS\" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY\n# KIND, either express or implied.  See the License for the\n# specific language governing permissions and limitations\n# under the License.\n#\n\n##\n## Edit a properties config file and replace values based on\n## the ENV variables\n## export my-key=new-value\n## ./apply-config-from-env file.conf\n##\n\nimport os, sys\n\nif len(sys.argv) < 2:\n    print('Usage: %s' % (sys.argv[0]))\n    sys.exit(1)\n\n# Always apply env config to env scripts as well\nconf_files = sys.argv[1:]\n\nPF_ENV_PREFIX = 'PULSAR_PREFIX_'\n\nfor conf_filename in conf_files:\n    lines = []  # List of config file lines\n    keys = {} # Map a key to its line number in the file\n\n    # Load conf file\n    for line in open(conf_filename):\n        lines.append(line)\n        line = line.strip()\n        if not line or line.startswith('#'):\n            continue\n\n        k,v = line.split('=', 1)\n        keys[k] = len(lines) - 1\n\n    # Update values from Env\n    for k in sorted(os.environ.keys()):\n        v = os.environ[k]\n        if k.startswith(PF_ENV_PREFIX):\n            k = k[len(PF_ENV_PREFIX):]\n        if k in keys:\n            print('[%s] Applying config %s = %s' % (conf_filename, k, v))\n            idx = keys[k]\n            lines[idx] = '%s=%s\\n' % (k, v)\n\n\n    # Add new keys from Env\n    for k in sorted(os.environ.keys()):\n        v = os.environ[k]\n        if not k.startswith(PF_ENV_PREFIX):\n            continue\n        k = k[len(PF_ENV_PREFIX):]\n        if k not in keys:\n            print('[%s] Adding config %s = %s' % (conf_filename, k, v))\n            lines.append('%s=%s\\n' % (k, v))\n        else:\n            print('[%s] Updating config %s = %s' %(conf_filename, k, v))\n            lines[keys[k]] = '%s=%s\\n' % (k, v)\n\n    # Store back the updated config in the same file\n    f = open(conf_filename, 'w')\n    for line in lines:\n        f.write(line)\n    f.close()"
