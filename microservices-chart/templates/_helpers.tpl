{{/*{{- define "service.fullname" -}}*/}}
{{/*{{ .Release.Name }}-{{ .Values.userService.name }}*/}}
{{/*{{- end -}}*/}}

{{/*{{- define "product.fullname" -}}*/}}
{{/*{{ .Release.Name }}-{{ .Values.productService.name }}*/}}
{{/*{{- end -}}*/}}

{{- define "service.fullname" -}}
{{ .Values.userService.name }}
{{- end -}}

{{- define "product.fullname" -}}
{{ .Values.productService.name }}
{{- end -}}