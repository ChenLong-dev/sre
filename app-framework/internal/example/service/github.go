package service

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/app-framework/internal/example/config"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/req"
	"gitlab.shanhai.int/sre/app-framework/internal/example/models/resp"
	"gitlab.shanhai.int/sre/library/base/null"
	"gitlab.shanhai.int/sre/library/goroutine"
	"gitlab.shanhai.int/sre/library/net/errcode"
	"gitlab.shanhai.int/sre/library/net/httpclient"
	"net/http"
	"strconv"
	"time"
)

func (s *Service) GetOwnerGithubAggregationResp(ctx context.Context, name string) (result *resp.GithubRepositoryAggregationResp, err error) {
	result = new(resp.GithubRepositoryAggregationResp)

	// ==========================
	// 获取redlock分布式锁
	// ==========================
	mutex := s.dao.GetOwnerGithubAggregationLock(ctx, name)
	// ==========================
	// 加锁
	// ==========================
	err = s.dao.LockOwnerGithub(ctx, mutex)
	if err != nil {
		return nil, err
	}
	// ==========================
	// defer解锁
	// ==========================
	defer func() {
		unlockError := s.dao.UnlockOwnerGithub(ctx, mutex)
		// ==========================
		// 当错误非空时，直接返回原始错误
		// ==========================
		if err != nil {
			return
		} else if unlockError != nil {
			// ==========================
			// 当错误为空，则使用闭包方式，替换返回值
			// ==========================
			err = unlockError
		}
	}()

	// ==========================
	// 声明协程组
	//
	// 通过 WithContext 声明协程组时
	// 协程组内任意协程返回错误/异常时，都会被捕获，并且在 Wait 函数返回值中返回首个错误
	// 在捕获错误的同时，也会取消协程组内其他协程，以避免资源浪费
	//
	// 如遇到无需取消其他协程(如，协程内的请求结果对该接口返回值重要性较低，允许降级)的场景
	// 请使用 New 声明协程组
	// ==========================
	wg := goroutine.WithContext(ctx, fmt.Sprintf("OwnerGithubAggreation-%s", name))

	//// New 方式声明协程组
	//wg := goroutine.New(fmt.Sprintf("OwnerGithubAggreation-%s", name))

	reposList := make([]*resp.GithubRepositoryResp, 0)
	// ==========================
	// 启动获取用户仓库协程
	// ==========================
	wg.Go(ctx, "repos", func(ctx context.Context) error {
		// ==========================
		// 注意，此处声明时不能使用外部名称，如reposList或err
		// 因为闭包函数内作用域与外部不同
		// 如果使用外部名称，闭包内部变量实际上是没有赋给外部的
		// ==========================
		list, e := s.GetOwnerGithubReposList(ctx, &req.GetGithubRepositoryListReq{
			Page:  null.IntFrom(1),
			Limit: null.IntFrom(5),
			Owner: name,
		})
		if e != nil {
			return e
		}

		// ==========================
		// 手动赋值变量给外部
		// ==========================
		reposList = list
		return nil
	})

	owner := new(resp.GithubOwnerResp)
	// ==========================
	// 启动获取用户信息协程
	// ==========================
	wg.Go(ctx, "detail", func(ctx context.Context) error {
		// ==========================
		// 注意，此处声明时不能使用外部名称，如reposList或err
		// 因为闭包函数内作用域与外部不同
		// 如果使用外部名称，闭包内部变量实际上是没有赋给外部的
		// ==========================
		res, e := s.GetOwnerGithubDetail(ctx, name)
		if e != nil {
			return e
		}

		// ==========================
		// 手动赋值变量给外部
		// ==========================
		owner = res
		return nil
	})

	// ==========================
	// 等待协程组内所有协程执行完毕
	//
	// 协程组内任意协程返回错误/异常时，都会被捕获
	// 并且在该函数返回值中返回首个错误
	// ==========================
	err = wg.Wait()
	// ==========================
	// 如遇到单个协程错误不影响其他协程(如，协程内的请求结果对该接口返回值重要性较低，允许降级)的场景
	// 该处判断非空时，不应直接返回结果，而是打印错误日志
	// ==========================
	if err != nil {
		// ==========================
		// 在首次生成error时，应当立即使用errors.Wrapf包裹
		// 外层只需直接返回error，无需再次包裹
		// ==========================
		return nil, err
	}

	// ==========================
	// 协程执行成功，赋值
	// ==========================
	result.Owner = owner
	result.Repos = reposList

	return result, nil
}

// 获取github用户仓库
func (s *Service) GetOwnerGithubReposList(ctx context.Context, getReq *req.GetGithubRepositoryListReq) ([]*resp.GithubRepositoryResp, error) {
	list := make([]*resp.GithubRepositoryResp, 0)

	// 调用http客户端获取数据
	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/users/%s/repos", config.Conf.Host.Github, getReq.Owner)).
		QueryParams(httpclient.NewUrlValue().Add("page", strconv.Itoa(getReq.Page.ValueOrZero())).
			Add("per_page", strconv.Itoa(getReq.Limit.ValueOrZero()))).
		Headers(httpclient.GetDefaultHeader()).
		RequestTimeout(time.Second * 15).
		// ==========================
		// 降级后的响应
		// ==========================
		DegradedJsonResponse(make([]*resp.GithubRepositoryResp, 0)).
		Fetch(ctx).
		DecodeJSON(&list)
	if err != nil {
		return nil, errors.Wrapf(errcode.InternalError, "%s", err)
	}

	return list, nil
}

// 获取github用户详细信息
func (s *Service) GetOwnerGithubDetail(ctx context.Context, name string) (*resp.GithubOwnerResp, error) {
	result := new(resp.GithubOwnerResp)

	// 调用http客户端获取数据
	err := s.httpClient.Builder().
		Method(http.MethodGet).
		URL(fmt.Sprintf("%s/users/%s", config.Conf.Host.Github, name)).
		Headers(httpclient.GetDefaultHeader()).
		RequestTimeout(time.Second * 15).
		// ==========================
		// 降级后的响应
		// ==========================
		DegradedJsonResponse(resp.GithubOwnerResp{
			Name: name,
		}).
		Fetch(ctx).
		DecodeJSON(&result)
	if err != nil {
		return nil, errors.Wrapf(errcode.InternalError, "%s", err)
	}

	return result, nil
}
