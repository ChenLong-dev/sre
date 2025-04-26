package resp

import (
	"rulai/models/entity"
)

type KongRouteInfoResp struct {
	KongName string   `json:"kong_name"`
	Hosts    []string `json:"hosts"`
	Paths    []string `json:"paths"`
}

type KongServiceResp struct {
	Next string `json:"next"`
	Data []*struct {
		ID       string   `json:"id"`
		Host     string   `json:"host"`
		Name     string   `json:"name"`
		Targets  []string `json:"targets"`
		KongHost string   `json:"kong_host"`
	}
}

type KongServiceRouteResp struct {
	Next string `json:"next"`
	Data []*struct {
		ID    string   `json:"id"`
		Paths []string `json:"paths"`
		Hosts []string `json:"hosts"`
	}
}

type KongUpstreamTargetResp struct {
	Next string `json:"next"`
	Data []*struct {
		ID     string `json:"id"`
		Target string `json:"target"`
	}
}

// GetKongUpstreamsResp 查询 Kong 网关 upstream 列表响应值
type GetKongUpstreamsResp struct {
	Next string                 `json:"next"`
	Data []*GetKongUpstreamResp `json:"data"`
}

// GetKongUpstreamResp 查询 Kong 网关 upstream 响应值
type GetKongUpstreamResp entity.KongUpstream

// GetKongTargetsResp 查询 Kong 网关 target 列表响应值
type GetKongTargetsResp struct {
	Next   string               `json:"next"`
	Data   []*GetKongTargetResp `json:"data"`
	Offset string               `json:"offset"`
}

// GetKongTargetResp 查询 Kong 网关 target 响应值
type GetKongTargetResp entity.KongTarget

// GetKongServiceResp 查询 Kong 网关 service 响应值
type GetKongServiceResp entity.KongService

// GetKongRouteResp 查询 Kong 网关 route 响应值
type GetKongRouteResp entity.KongRoute

// GetKongServicesResp 查询 Kong 网关 services 列表
type GetKongServicesResp struct {
	Next string                `json:"next"`
	Data []*GetKongServiceResp `json:"data"`
}
