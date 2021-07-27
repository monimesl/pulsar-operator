package pulsarcluster

import (
	"fmt"
	"github.com/monimesl/operator-helper/oputil"
	"github.com/prometheus/common/log"
	v1 "k8s.io/api/core/v1"
	"strings"
)

const (
	pulsarConfigEnvPrefix = "PULSAR_PREFIX_"
)

var excludedOptions = addPulsarEnvPrefix([]string{
	"clusterName", "zookeeperServers",
	"configurationStoreServers", "PULSAR_GC",
	"PULSAR_MEM", "PULSAR_EXTRA_OPTS", "PULSAR_GC_LOG",
})

func processEnvVars(envs []v1.EnvVar) []v1.EnvVar {
	newEnvs := make([]v1.EnvVar, 0)
	for i := range envs {
		env := envs[i]
		name := env.Name
		actual := name
		if !strings.HasPrefix(name, pulsarConfigEnvPrefix) {
			name = fmt.Sprintf("%s%s", pulsarConfigEnvPrefix, name)
		}
		if oputil.Contains(excludedOptions, name) {
			log.Warnf("ignoring the config: %s", actual)
			continue
		}
		env.Name = name
		newEnvs[i] = env
		newEnvs = append(newEnvs, env)
	}
	return newEnvs
}

func processEnvVarMap(envs map[string]string) map[string]string {
	newEnvs := map[string]string{}
	for name, v := range envs {
		actual := name
		if !strings.HasPrefix(name, pulsarConfigEnvPrefix) {
			name = fmt.Sprintf("%s%s", pulsarConfigEnvPrefix, name)
		}
		if oputil.Contains(excludedOptions, name) {
			log.Warnf("ignoring the config: %s", actual)
			continue
		}
		newEnvs[name] = v
	}
	return newEnvs
}

func addPulsarEnvPrefix(envs []string) []string {
	newEnvs := make([]string, len(envs))
	for i := range envs {
		env := envs[i]
		if !strings.HasPrefix(env, pulsarConfigEnvPrefix) {
			env = fmt.Sprintf("%s%s", pulsarConfigEnvPrefix, env)
		}
		newEnvs[i] = env
	}
	return envs
}
