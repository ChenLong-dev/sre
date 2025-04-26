max_over_time (
    (sum (
        container_memory_usage_bytes {
            {{.ContainerLabelName}}=~"{{.ContainerName}}",
            image!="",
            job="kubelet",
            namespace=~"{{.EnvName}}"
        }
    ) / sum (
        kube_pod_container_resource_requests{
            namespace=~"{{.EnvName}}",
            resource="memory", 
            {{.ContainerLabelName}}=~"{{.ContainerName}}"
    })) [{{.CountTime}}:5m]
) < {{.UsageRateLimit}}