package service

import (
	"rulai/models/entity"
	"rulai/utils"

	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/pkg/errors"
	"gitlab.shanhai.int/sre/library/log"
	"gitlab.shanhai.int/sre/library/net/errcode"
)

// RenderK8sTemplate 渲染k8s模板(增加集群区分)
func (s *Service) RenderK8sTemplate(ctx context.Context, relativeDir string,
	clusterName entity.ClusterName,
	env entity.AppEnvName, templateReq entity.K8sObjectTemplate) (string, error) {
	clusterInfo, err := s.getClusterInfo(clusterName, string(env))
	if err != nil {
		return "", err
	}

	workloadTemplateReq, ok := templateReq.(entity.K8sWorkloadObjectTemplate)
	if ok {
		workloadTemplateReq.UnifyImageName(clusterInfo.imageRegistryHostWithNamespace)
	}

	kind := templateReq.Kind()
	groupVersion, err := s.getK8sResourceGroupVersionFromClusterInfo(clusterInfo, kind)
	if err != nil {
		return "", err
	}

	templateReq.SetAPIVersion(groupVersion.String())
	filepath := fmt.Sprintf("%s%s.yaml", relativeDir, kind)

	// 日志输出模板信息
	log.Infoc(ctx, "filepath:%v clusterName:%v env:%v templateReq:%v",
		filepath, clusterName, env, templateReq)
	// 检测是否指定集群有特殊模板文件(文件名带有集群名称)
	clusterVersion, err := s.getClusterVersion(ctx, clusterName, env)
	if err != nil {
		log.Errorc(ctx, "Cluster(%s) not exists ... fallback to base template", clusterName)
		return s.RenderTemplate(ctx, filepath, templateReq)
	}

	dir, filename := path.Split(filepath)
	ext := path.Ext(filename)
	baseFilename := filename[:len(filename)-len(ext)]
	minorClusterVersion := utils.UnifyK8sMinorVersion(clusterVersion.Minor)
	newPath := fmt.Sprintf("%s%s-%s-%s%s", dir, baseFilename, clusterVersion.Major, minorClusterVersion, ext)
	if _, err = os.Stat(newPath); err == nil {
		return s.RenderTemplate(ctx, newPath, templateReq)
	}

	data, err := s.RenderTemplate(ctx, filepath, templateReq)
	// 日志输出渲染后yaml文件
	log.Infoc(ctx, "rendered yaml:\n%v\n", data)
	return data, err
}

// RenderTemplate 渲染模版
func (s *Service) RenderTemplate(_ context.Context, filepath string, templateReq interface{}) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return "", errors.Wrapf(errcode.InternalError, "open file error:%s", err.Error())
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(f)
	if err != nil {
		return "", errors.Wrapf(errcode.InternalError, "read template error:%s", err.Error())
	}

	tpl, err := template.New(filepath).Parse(buf.String())
	if err != nil {
		return "", errors.Wrapf(errcode.InternalError, "parse template error:%s", err.Error())
	}

	buf.Reset()
	err = tpl.Execute(buf, templateReq)
	if err != nil {
		return "", errors.Wrapf(errcode.InternalError, "render template error:%s", err.Error())
	}

	return buf.String(), nil
}

// RenderTemplateFromText render template from text
func (s *Service) RenderTemplateFromText(templateReq interface{}, tmpName, parseText string) (string, error) {
	buf := new(bytes.Buffer)
	tmpl, err := template.New(tmpName).Parse(parseText)
	if err != nil {
		return "", errors.Wrapf(errcode.InternalError, "parse template error:%s", err.Error())
	}
	if err := tmpl.Execute(buf, templateReq); err != nil {
		return "", errors.Wrapf(errcode.InternalError, "parse template error:%s", err.Error())
	}

	return buf.String(), nil
}
