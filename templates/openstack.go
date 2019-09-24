package templates

const OpenStackCloudProviderTemplate = `[Global]
{{- if ne .OpenstackConfig.Global.AuthURL "" }}
auth-url = "{{ .OpenstackConfig.Global.AuthURL }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.Username "" }}
username = "{{ .OpenstackConfig.Global.Username }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.UserID "" }}
user-id = "{{ .OpenstackConfig.Global.UserID }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.Password "" }}
password = "{{ .OpenstackConfig.Global.Password }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.TenantID "" }}
tenant-id = "{{ .OpenstackConfig.Global.TenantID }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.TenantName "" }}
tenant-name = "{{ .OpenstackConfig.Global.TenantName }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.TrustID "" }}
trust-id = "{{ .OpenstackConfig.Global.TrustID }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.DomainID "" }}
domain-id = "{{ .OpenstackConfig.Global.DomainID }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.DomainName "" }}
domain-name = "{{ .OpenstackConfig.Global.DomainName }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.Region "" }}
region = "{{ .OpenstackConfig.Global.Region }}"
{{- end }}
{{- if ne .OpenstackConfig.Global.CAFile "" }}
ca-file = "{{ .OpenstackConfig.Global.CAFile }}"
{{- end }}

[LoadBalancer]
{{- if ne .OpenstackConfig.LoadBalancer.LBVersion "" }}
lb-version = "{{ .OpenstackConfig.LoadBalancer.LBVersion }}"
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.UseOctavia false }}
use-octavia = {{ .OpenstackConfig.LoadBalancer.UseOctavia }}
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.SubnetID "" }}
subnet-id = "{{ .OpenstackConfig.LoadBalancer.SubnetID }}"
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.FloatingNetworkID "" }}
floating-network-id = "{{ .OpenstackConfig.LoadBalancer.FloatingNetworkID }}"
{{- end }}
{{- if (avail "FloatingSubnetID" .OpenstackConfig.LoadBalancer) }}
{{- if ne .OpenstackConfig.LoadBalancer.FloatingSubnetID "" }}
floating-subnet-id = "{{ .OpenstackConfig.LoadBalancer.FloatingSubnetID }}"
{{- end }}
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.LBMethod "" }}
lb-method = "{{ .OpenstackConfig.LoadBalancer.LBMethod }}"
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.LBProvider "" }}
lb-provider = "{{ .OpenstackConfig.LoadBalancer.LBProvider }}"
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.CreateMonitor false }}
create-monitor = {{ .OpenstackConfig.LoadBalancer.CreateMonitor }}
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.MonitorDelay "" }}
monitor-delay = "{{ .OpenstackConfig.LoadBalancer.MonitorDelay }}"
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.MonitorTimeout "" }}
monitor-timeout = "{{ .OpenstackConfig.LoadBalancer.MonitorTimeout }}"
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.MonitorMaxRetries 0 }}
monitor-max-retries = {{ .OpenstackConfig.LoadBalancer.MonitorMaxRetries }}
{{- end }}
{{- if ne .OpenstackConfig.LoadBalancer.ManageSecurityGroups false }}
manage-security-groups = {{ .OpenstackConfig.LoadBalancer.ManageSecurityGroups }}
{{- end }}

{{- if (avail "LBClasses" .OpenstackConfig.LoadBalancer) }}
{{- range $k,$v := .OpenstackConfig.LoadBalancer.LBClasses }}

[LoadBalancerClass "{{ $k }}"]
{{- if ne $v.FloatingNetworkID "" }}
floating-network-id = "{{ $v.FloatingNetworkID }}"
{{- end }}
{{- if ne $v.FloatingSubnetID "" }}
floating-subnet-id = "{{ $v.FloatingSubnetID }}"
{{- end }}
{{- if ne $v.SubnetID "" }}
subnet-id = "{{ $v.SubnetID }}"
{{- end }}
{{- end }}
{{- end }}

[BlockStorage]
{{- if ne .OpenstackConfig.BlockStorage.BSVersion "" }}
bs-version = "{{ .OpenstackConfig.BlockStorage.BSVersion }}"
{{- end }}
{{- if ne .OpenstackConfig.BlockStorage.TrustDevicePath false }}
trust-device-path = {{ .OpenstackConfig.BlockStorage.TrustDevicePath }}
{{- end }}
{{- if ne .OpenstackConfig.BlockStorage.IgnoreVolumeAZ false }}
ignore-volume-az = {{ .OpenstackConfig.BlockStorage.IgnoreVolumeAZ }}
{{- end }}

[Route]
{{- if ne .OpenstackConfig.Route.RouterID "" }}
router-id = "{{ .OpenstackConfig.Route.RouterID }}"
{{- end }}

[Metadata]
{{- if ne .OpenstackConfig.Metadata.SearchOrder "" }}
search-order = "{{ .OpenstackConfig.Metadata.SearchOrder }}"
{{- end }}
{{- if ne .OpenstackConfig.Metadata.RequestTimeout 0 }}
request-timeout = {{ .OpenstackConfig.Metadata.RequestTimeout }}
{{- end }}
`
