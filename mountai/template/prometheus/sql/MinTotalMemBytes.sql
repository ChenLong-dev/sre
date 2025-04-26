min_over_time (
    sum by (namespace, {{.ContainerLabelName}}) (
        container_memory_usage_bytes {
            {{.ContainerLabelName}}=~"{{.ContainerName}}",
            image!="",
            job="kubelet",
            namespace=~"{{.EnvName}}"
        }
    ) [{{.CountTime}}:5m]
)