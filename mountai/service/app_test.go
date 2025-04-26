package service

import (
	"rulai/models/entity"
	"rulai/models/req"

	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_CalculateAppRecommendResource(t *testing.T) {
	t.Run("commodity-subscription", func(t *testing.T) {
		res, err := s.CalculateAppRecommendResource(context.Background(), &req.CalculateAppRecommendReq{
			EnvName:           entity.AppEnvPrd,
			DailyMaxTotalCPU:  0.7417099395335687,
			WeeklyMaxTotalCPU: 2.3003444101292505,
			DailyMaxTotalMem:  65699840,
			WeeklyMaxTotalMem: 187486208,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("user-system", func(t *testing.T) {
		res, err := s.CalculateAppRecommendResource(context.Background(), &req.CalculateAppRecommendReq{
			EnvName:           entity.AppEnvPrd,
			DailyMaxTotalCPU:  2.8367391816332845,
			WeeklyMaxTotalCPU: 2.9597931500995847,
			DailyMaxTotalMem:  341172224,
			WeeklyMaxTotalMem: 3761094656,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("app-framework", func(t *testing.T) {
		res, err := s.CalculateAppRecommendResource(context.Background(), &req.CalculateAppRecommendReq{
			EnvName:           entity.AppEnvPrd,
			DailyMaxTotalCPU:  0.0029373371851862534,
			WeeklyMaxTotalCPU: 0.0034696873481481014,
			DailyMaxTotalMem:  17260544,
			WeeklyMaxTotalMem: 154419200,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("web-bff", func(t *testing.T) {
		res, err := s.CalculateAppRecommendResource(context.Background(), &req.CalculateAppRecommendReq{
			EnvName:           entity.AppEnvPrd,
			DailyMaxTotalCPU:  0.36280639709625073,
			WeeklyMaxTotalCPU: 0.5360272742332941,
			DailyMaxTotalMem:  3442814976,
			WeeklyMaxTotalMem: 4095463424,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("vip-subscription", func(t *testing.T) {
		res, err := s.CalculateAppRecommendResource(context.Background(), &req.CalculateAppRecommendReq{
			EnvName:           entity.AppEnvPrd,
			DailyMaxTotalCPU:  3.1297064677570483,
			WeeklyMaxTotalCPU: 3.1297064677570483,
			DailyMaxTotalMem:  170749952,
			WeeklyMaxTotalMem: 170749952,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})

	t.Run("pop-api", func(t *testing.T) {
		res, err := s.CalculateAppRecommendResource(context.Background(), &req.CalculateAppRecommendReq{
			EnvName:           entity.AppEnvPrd,
			DailyMaxTotalCPU:  5.117129173423592,
			WeeklyMaxTotalCPU: 5.117129173423592,
			DailyMaxTotalMem:  1337647104,
			WeeklyMaxTotalMem: 1337647104,
		})
		assert.Nil(t, err)
		fmt.Printf("%#v\n", res)
	})
}
