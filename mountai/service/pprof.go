package service

import (
	"rulai/models/entity"
	"rulai/models/req"

	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/pprof/driver"
	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// GenerateAnalyzedPProfFile 生成分析后的pprof文件
func (s *Service) GenerateAnalyzedPProfFile(ctx context.Context, createReq *req.CreatePProfReq) error {
	// 生成目录
	err := os.MkdirAll(filepath.Dir(createReq.GenerateFilePath), 0755)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}

	analysedCommandList := []string{
		fmt.Sprintf("output=%s", createReq.GenerateFilePath),
		string(createReq.Action),
		"exit",
	}

	// 分析文件
	log.Infoc(ctx, "analyze pprof file")

	if err := driver.PProf(&driver.Options{
		UI: &entity.PProfAnalyzedUI{
			CommandIndex: 0,
			CommandList:  analysedCommandList,
		},
		Flagset: &entity.PProfAnalyzedFlags{
			FlagSet:     flag.NewFlagSet(os.Args[0], flag.ExitOnError),
			CommandList: []string{createReq.SourceFilePath},
		},
	}); err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}
	return nil
}

// GenerateOriginPProfFile 生成原始的pprof文件
func (s *Service) GenerateOriginPProfFile(ctx context.Context, createReq *req.CreatePProfReq) error {
	// 检查是否安装curl
	log.Infoc(ctx, "check curl command")
	_, err := s.ExecPodCommand(context.Background(), createReq.ClusterName,
		&req.ExecPodReq{
			Namespace: createReq.Namespace,
			Name:      createReq.PodName,
			Env:       string(createReq.EnvName),
			Commands:  []string{"curl", "--help"},
			Container: createReq.Container,
		})
	// 不存在curl command
	if err != nil && strings.Contains(err.Error(), "exit code 126") {
		// 替换apk源
		log.Infoc(ctx, "replace apk resource")
		_, e := s.ExecPodCommand(ctx, createReq.ClusterName,
			&req.ExecPodReq{
				Namespace: createReq.Namespace,
				Name:      createReq.PodName,
				Env:       string(createReq.EnvName),
				Container: createReq.Container,
				Commands: []string{
					"sed", "-i", "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g", "/etc/apk/repositories",
				},
			})
		if e != nil {
			return e
		}

		// 安装curl
		log.Infoc(ctx, "install curl")
		_, err = s.ExecPodCommand(ctx, createReq.ClusterName,
			&req.ExecPodReq{
				Namespace: createReq.Namespace,
				Env:       string(createReq.EnvName),
				Name:      createReq.PodName,
				Container: createReq.Container,
				Commands: []string{
					"apk", "add", "-U", "curl",
				},
			})
		if err != nil {
			return err
		}
	}

	var params string
	if createReq.Seconds != 0 {
		params = fmt.Sprintf("seconds=%d&", createReq.Seconds)
	}

	// 请求pprof接口
	log.Infoc(ctx, "get origin pprof file")
	data, err := s.ExecPodCommand(ctx, createReq.ClusterName,
		&req.ExecPodReq{
			Namespace: createReq.Namespace,
			Env:       string(createReq.EnvName),
			Name:      createReq.PodName,
			Container: createReq.Container,
			Commands: []string{
				"curl", "-s",
				fmt.Sprintf("http://localhost:%d/debug/pprof/%s?%s", createReq.PodPort, createReq.Type, params),
			},
		})
	if err != nil {
		return err
	}

	// 生成文件
	log.Infoc(ctx, "create origin pprof file")
	err = os.MkdirAll(filepath.Dir(createReq.SourceFilePath), 0755)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}
	originFile, err := os.Create(createReq.SourceFilePath)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}
	defer originFile.Close()
	_, err = originFile.Write(data)
	if err != nil {
		return errors.Wrap(errcode.InternalError, err.Error())
	}

	return nil
}
