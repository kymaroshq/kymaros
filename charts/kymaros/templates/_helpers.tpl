{{/*
Expand the name of the chart.
*/}}
{{- define "kymaros.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kymaros.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart label value.
*/}}
{{- define "kymaros.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels applied to all resources.
*/}}
{{- define "kymaros.labels" -}}
helm.sh/chart: {{ include "kymaros.chart" . }}
{{ include "kymaros.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels for the controller.
*/}}
{{- define "kymaros.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kymaros.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Selector labels for the API server.
*/}}
{{- define "kymaros.apiSelectorLabels" -}}
app.kubernetes.io/name: {{ include "kymaros.name" . }}-api
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Controller service account name.
*/}}
{{- define "kymaros.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (printf "%s-controller" (include "kymaros.fullname" .)) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
API server service account name.
*/}}
{{- define "kymaros.apiServiceAccountName" -}}
{{- if .Values.serviceAccountApi.create }}
{{- default (printf "%s-api" (include "kymaros.fullname" .)) .Values.serviceAccountApi.name }}
{{- else }}
{{- default "default" .Values.serviceAccountApi.name }}
{{- end }}
{{- end }}

{{/*
Controller image reference.
*/}}
{{- define "kymaros.controllerImage" -}}
{{- $tag := .Values.controller.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.controller.image.repository $tag }}
{{- end }}

{{/*
API server image reference.
*/}}
{{- define "kymaros.apiImage" -}}
{{- $tag := .Values.api.image.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.api.image.repository $tag }}
{{- end }}

{{/*
Global imagePullSecrets merged with component-specific ones.
*/}}
{{- define "kymaros.imagePullSecrets" -}}
{{- with .Values.global.imagePullSecrets }}
imagePullSecrets:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}
