min_over_time (
    sum by (namespace, {{.ContainerLabelName}}) (
        rate (
            container_cpu_usage_seconds_total {
                {{.ContainerLabelName}}=~"{{.ContainerName}}",
                image!="",
                job="kubelet",
                namespace=~"{{.EnvName}}"
            } [5m]
        )
    ) [{{.CountTime}}:5m]
)