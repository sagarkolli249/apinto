{{- define "apipark.name" -}}apipark{{- end -}}
{{- define "apipark.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "apipark.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "apipark.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
app.kubernetes.io/name: {{ include "apipark.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
