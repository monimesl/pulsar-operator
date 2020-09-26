{{/*
Expand the name of the chart.
*/}}
{{- define "pulsar-operator-chart.name" -}}
{{ .Chart.Name }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "pulsar-operator-chart.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "pulsar-operator-chart.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "pulsar-operator-chart.labels" -}}
helm.sh/chart: {{ include "pulsar-operator-chart.chart" . }}
{{ include "pulsar-operator-chart.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "pulsar-operator-chart.selectorLabels" -}}
app.kubernetes.io/name: {{ include "pulsar-operator-chart.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
