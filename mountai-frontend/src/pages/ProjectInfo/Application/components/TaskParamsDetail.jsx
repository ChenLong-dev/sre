import React, { useState, useEffect, useMemo } from 'react';
import { Descriptions, Modal } from 'antd';
import PropTypes from 'prop-types';
import { includes, keyBy } from 'lodash';
import ReactJson from 'react-json-view';
import {
  CRON_JOB,
  ONE_TIME_JOB,
  SERVICE,
  WORKER,
} from '@/constants/applicationTypes';
import { getLatestTasks } from '../service';

const TaskParamsDetail = ({
  onHide,
  application,
  filterInfo,
  running,
  clusterName,
  projectInfo,
}) => {
  const [taskParams, setTaskParams] = useState({});

  useEffect(() => {
    (async () => {
      const latestTasks = await getLatestTasks(
        application.id,
        filterInfo?.envname,
        clusterName,
        running.version,
        true
      );
      setTaskParams(latestTasks?.param ?? {});
    })();
  }, []);

  const configRenameModes = useMemo(
    () => keyBy(projectInfo.config_rename_modes, 'enum'),
    []
  );

  return (
    <Modal
      width={800}
      centered
      title="任务参数详情"
      visible
      onCancel={onHide}
      onOk={onHide}
    >
      <Descriptions column={1}>
        <Descriptions.Item label="应用名称">
          {application?.name}
        </Descriptions.Item>
        <Descriptions.Item label="环境">
          {filterInfo?.envname}
        </Descriptions.Item>
        <Descriptions.Item label="部署版本">
          {running.version}
        </Descriptions.Item>
        <Descriptions.Item label="镜像">
          {taskParams.image_version}
        </Descriptions.Item>
        <Descriptions.Item label="创建时间">
          {running.create_time}
        </Descriptions.Item>
        {taskParams.config_commit_id && (
          <Descriptions.Item label="配置ID">
            {taskParams.config_commit_id}
          </Descriptions.Item>
        )}
        {taskParams.config_rename_prefix && (
          <Descriptions.Item label="特殊配置重命名前缀">
            {taskParams.config_rename_prefix}
          </Descriptions.Item>
        )}
        {taskParams.config_rename_mode && (
          <Descriptions.Item label="特殊配置重命名模式">
            {configRenameModes?.[taskParams.config_rename_mode]?.name ??
              taskParams.config_rename_mode}
          </Descriptions.Item>
        )}
        {includes([CRON_JOB, ONE_TIME_JOB], application.type) && (
          <>
            <Descriptions.Item label="执行命令">
              {application.type === CRON_JOB
                ? taskParams.cron_command
                : taskParams.job_command}
            </Descriptions.Item>
            <Descriptions.Item label="执行任务超时时间">
              {taskParams.active_deadline_seconds}
            </Descriptions.Item>
            <Descriptions.Item label="调度周期">
              {taskParams.cron_param}
            </Descriptions.Item>
            <Descriptions.Item label="失败重试次数">
              {taskParams.backoff_limit}
            </Descriptions.Item>
          </>
        )}
        {application.type === SERVICE && (
          <>
            <Descriptions.Item label="健康检查">
              {taskParams.health_check_url}
            </Descriptions.Item>
            <Descriptions.Item label="端口">
              {taskParams.target_port}
            </Descriptions.Item>
            <Descriptions.Item label="额外端口">
              <ReactJson
                src={taskParams.exposed_ports || {}}
                displayDataTypes={false}
                displayObjectSize={false}
                name={null}
              />
            </Descriptions.Item>
          </>
        )}
        <Descriptions.Item label="容器CPU规格">
          {taskParams.cpu_request}
        </Descriptions.Item>
        <Descriptions.Item label="容器内存规格">
          {taskParams.mem_request}
        </Descriptions.Item>
        {!includes([CRON_JOB, ONE_TIME_JOB], application.type) && (
          <>
            <Descriptions.Item label="实例数">
              {taskParams.min_pod_count}
            </Descriptions.Item>
            <Descriptions.Item label="自动伸缩">
              {taskParams.is_auto_scale ? '是' : '否'}
            </Descriptions.Item>
            {taskParams.is_auto_scale && (
              <Descriptions.Item label="最大实例数">
                {taskParams.max_pod_count}
              </Descriptions.Item>
            )}
          </>
        )}
        <Descriptions.Item label="环境变量">
          <ReactJson
            src={taskParams.vars || {}}
            displayDataTypes={false}
            displayObjectSize={false}
            name={null}
          />
        </Descriptions.Item>
        {!includes([CRON_JOB, ONE_TIME_JOB], application.type) && (
          <>
            {application.type === SERVICE && (
              <Descriptions.Item label="采集Metrics">
                {taskParams.is_support_metrics ? '是' : '否'}
              </Descriptions.Item>
            )}
            <Descriptions.Item label="启动命令">
              {taskParams.cover_command}
            </Descriptions.Item>
            <Descriptions.Item label="预执行命令">
              {taskParams.pre_stop_command}
            </Descriptions.Item>
          </>
        )}
        <Descriptions.Item
          label="高级选项"
          labelStyle={{ fontWeight: 'bold' }}
        />
        <Descriptions.Item label="容器CPU限制">
          {taskParams.cpu_limit}
        </Descriptions.Item>
        <Descriptions.Item label="容器内存限制">
          {taskParams.mem_limit}
        </Descriptions.Item>
        <Descriptions.Item label="服务分级规格">
          {taskParams.node_affinity_label_config?.importance}
        </Descriptions.Item>
        <Descriptions.Item label="节点cpu规格">
          {taskParams.node_affinity_label_config?.cpu}
        </Descriptions.Item>
        <Descriptions.Item label="节点内存规格">
          {taskParams.node_affinity_label_config?.mem}
        </Descriptions.Item>
        {includes([SERVICE, WORKER, ONE_TIME_JOB], application.type) && (
          <Descriptions.Item label="专用标记">
            {taskParams.node_affinity_label_config?.exclusive}
          </Descriptions.Item>
        )}
        <Descriptions.Item label="优雅终止时长">{`${taskParams.termination_grace_period_sec}秒`}</Descriptions.Item>
        <Descriptions.Item label="开启会话保持">
          {taskParams.is_support_sticky_session ? '是' : '否'}
        </Descriptions.Item>
        {taskParams.is_support_sticky_session && (
          <Descriptions.Item label="会话保持时间">{`${taskParams.session_cookie_max_age}秒`}</Descriptions.Item>
        )}
      </Descriptions>
    </Modal>
  );
};

TaskParamsDetail.propTypes = {
  onHide: PropTypes.func.isRequired,
  application: PropTypes.object.isRequired,
  filterInfo: PropTypes.object.isRequired,
  running: PropTypes.object.isRequired,
};

export default React.memo(TaskParamsDetail);
