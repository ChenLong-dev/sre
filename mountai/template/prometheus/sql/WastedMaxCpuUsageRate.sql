max_over_time (
    (sum (rate (
        container_cpu_usage_seconds_total {
            {{.ContainerLabelName}}=~"{{.ContainerName}}",
            image!="",
            job="kubelet",
            namespace=~"{{.EnvName}}"
        } [5m]
    )) / sum (
        kube_pod_container_resource_requests{
            namespace=~"{{.EnvName}}",
            resource="cpu", 
            {{.ContainerLabelName}}=~"{{.ContainerName}}"
    })) [{{.CountTime}}:5m]
) < {{.UsageRateLimit}}