{{/*
Return the name of the apinto chart
*/}}
{{- define "apinto.name" -}}
apinto
{{- end -}}

{{/*
Return the fully qualified name: <release>-apinto
*/}}
{{- define "apinto.fullname" -}}
{{ printf "%s-%s" .Release.Name (include "apinto.name" .) }}
{{- end -}}

{{/*
Standard labels for apinto resources
*/}}
{{- define "apinto.labels" -}}
app.kubernetes.io/name: {{ include "apinto.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/component: gateway
app.kubernetes.io/part-of: apipark
{{- end -}}

{{/*
Selector labels for pod selectors
*/}}
{{- define "apinto.selectorLabels" -}}
app.kubernetes.io/name: {{ include "apinto.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}