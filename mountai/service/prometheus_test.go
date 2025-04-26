package service

import (
	"rulai/models/entity"
	"rulai/models/req"

	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetPrometheusData(t *testing.T) {
	t.Run("max_total_cpu", func(t *testing.T) {
		data, err := s.RenderTemplate(context.Background(),
			"../template/prometheus/sql/MaxTotalCpuTime.sql", &entity.MaxTotalCPUTemplate{
				EnvName:            "stg",
				ContainerName:      "qt.*",
				CountTime:          "3d",
				ContainerLabelName: "container",
			})
		assert.Nil(t, err)
		fmt.Printf("%s\n", data)

		_, err = s.GetPrometheusData(context.Background(), &req.QueryPrometheusReq{
			EnvName: "stg",
			SQL:     data,
		})
		assert.Nil(t, err)
	})

	t.Run("max_total_mem", func(t *testing.T) {
		data, err := s.RenderTemplate(context.Background(),
			"../template/prometheus/sql/MaxTotalMemBytes.sql", &entity.MaxTotalMemTemplate{
				EnvName:            "stg",
				ContainerName:      "qt.*",
				CountTime:          "3d",
				ContainerLabelName: "container",
			})
		assert.Nil(t, err)
		fmt.Printf("%s\n", data)

		_, err = s.GetPrometheusData(context.Background(), &req.QueryPrometheusReq{
			EnvName: "stg",
			SQL:     data,
		})
		assert.Nil(t, err)
	})

	t.Run("wasted_max_cpu_usage_rate", func(t *testing.T) {
		data, err := s.RenderTemplate(context.Background(),
			"../template/prometheus/sql/WastedMaxCpuUsageRate.sql", &entity.WastedMaxCPUUsageRateTemplate{
				EnvName:            "stg",
				ContainerLabelName: "container",
				ContainerName:      "qt.*",
				CountTime:          "3d",
				UsageRateLimit:     0.1,
				MinCPUResource:     "0.1",
			})
		assert.Nil(t, err)
		fmt.Printf("%s\n", data)

		_, err = s.GetPrometheusData(context.Background(), &req.QueryPrometheusReq{
			EnvName: "stg",
			SQL:     data,
		})
		assert.Nil(t, err)
	})

	t.Run("wasted_max_mem_usage_rate", func(t *testing.T) {
		data, err := s.RenderTemplate(context.Background(),
			"../template/prometheus/sql/WastedMaxMemUsageRate.sql", &entity.WastedMaxMemUsageRateTemplate{
				EnvName:            "stg",
				ContainerLabelName: "container",
				ContainerName:      "qt.*",
				CountTime:          "3d",
				UsageRateLimit:     0.1,
				MinMemResource:     "268435456",
			})
		assert.Nil(t, err)
		fmt.Printf("%s\n", data)

		_, err = s.GetPrometheusData(context.Background(), &req.QueryPrometheusReq{
			EnvName: "stg",
			SQL:     data,
		})
		assert.Nil(t, err)
	})
}
