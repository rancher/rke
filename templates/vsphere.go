package templates

const VsphereCloudProviderTemplate = `
[Global]
{{- if and (ne .VsphereConfig.Global.SecretName "") (ne .VsphereConfig.Global.SecretNamespace "") }}
secret-name = "{{ .VsphereConfig.Global.SecretName }}"
secret-namespace = "{{ .VsphereConfig.Global.SecretNamespace }}"
{{- else }}
user = "{{ .VsphereConfig.Global.User }}"
password = "{{ .VsphereConfig.Global.Password }}"
{{- end }}
{{- if ne .VsphereConfig.Global.VCenterIP "" }}
server = "{{ .VsphereConfig.Global.VCenterIP }}"
{{- end }}
{{- if ne .VsphereConfig.Global.VCenterPort "" }}
port = "{{ .VsphereConfig.Global.VCenterPort }}"
{{- end }}
insecure-flag = "{{ .VsphereConfig.Global.InsecureFlag }}"
{{- if ne .VsphereConfig.Global.Datacenters "" }}
datacenters = "{{ .VsphereConfig.Global.Datacenters }}"
{{- end }}
{{- if ne .VsphereConfig.Global.Datacenter "" }}
datacenter = "{{ .VsphereConfig.Global.Datacenter }}"
{{- end }}
{{- if ne .VsphereConfig.Global.DefaultDatastore "" }}
datastore = "{{ .VsphereConfig.Global.DefaultDatastore }}"
{{- end }}
{{- if ne .VsphereConfig.Global.WorkingDir "" }}
working-dir = "{{ .VsphereConfig.Global.WorkingDir }}"
{{- end }}
soap-roundtrip-count = "{{ .VsphereConfig.Global.RoundTripperCount }}"
{{- if ne .VsphereConfig.Global.VMUUID "" }}
vm-uuid = "{{ .VsphereConfig.Global.VMUUID }}"
{{- end }}
{{- if ne .VsphereConfig.Global.VMName "" }}
vm-name = "{{ .VsphereConfig.Global.VMName }}"
{{- end }}

{{ range $k,$v := .VsphereConfig.VirtualCenter }}
[VirtualCenter "{{ $k }}"]
        {{- if and (ne $v.SecretName "") (ne $v.SecretNamespace "") }}
        secret-name = "{{ $v.SecretName }}"
        secret-namespace = "{{ $v.SecretNamespace }}"
        {{- else }}
        user = "{{ $v.User }}"
        password = "{{ $v.Password }}"
        {{- end }}
        {{- if ne $v.VCenterPort "" }}
        port = "{{ $v.VCenterPort }}"
        {{- end }}
        {{- if ne $v.Datacenters "" }}
        datacenters = "{{ $v.Datacenters }}"
        {{- end }}
        soap-roundtrip-count = "{{ $v.RoundTripperCount }}"
{{- end }}

[Workspace]
        server = "{{ .VsphereConfig.Workspace.VCenterIP }}"
        datacenter = "{{ .VsphereConfig.Workspace.Datacenter }}"
        folder = "{{ .VsphereConfig.Workspace.Folder }}"
        default-datastore = "{{ .VsphereConfig.Workspace.DefaultDatastore }}"
        resourcepool-path = "{{ .VsphereConfig.Workspace.ResourcePoolPath }}"

[Disk]
        {{- if ne .VsphereConfig.Disk.SCSIControllerType "" }}
        scsicontrollertype = {{ .VsphereConfig.Disk.SCSIControllerType }}
        {{- end }}

[Network]
        {{- if ne .VsphereConfig.Network.PublicNetwork "" }}
        public-network = "{{ .VsphereConfig.Network.PublicNetwork }}"
        {{- end }}
`
