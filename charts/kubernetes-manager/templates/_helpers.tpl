{{- define "env" -}}
- name: POD_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: POD_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
{{ if .Values.jaeger.agentOnHost }}
- name: JAEGER_AGENT_HOST
  valueFrom:
    fieldRef:
      apiVersion: v1
      fieldPath: status.hostIP
{{ end }}
{{ if .Values.sentry.enabled }}
- name: "SENTRY_DSN"
  value: {{ .Values.sentry.host }}
{{ end }}
{{ if .Values.env }}
{{ toYaml .Values.env }}
{{ end }}
{{- end -}}

{{- define "args" -}}
- --front.dist=/app/dist
- --config=/config/config.yaml
{{ range .Values.args }}
- {{ . }}
{{ end }}
{{- end -}}