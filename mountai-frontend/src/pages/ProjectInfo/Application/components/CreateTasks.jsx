import React, { useState, useEffect, useMemo } from 'react';
import {
  Form,
  Button,
  Input,
  Modal,
  Tooltip,
  Select,
  Switch,
  InputNumber,
  Space,
  Cascader,
  Checkbox,
  Alert,
} from 'antd';
import { map, forIn, forEach, sortBy, isArray } from 'lodash';

import {
  PlusOutlined,
  MinusCircleOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import {
  CRON_JOB,
  ONE_TIME_JOB,
  SERVICE,
  WORKER,
} from '@/constants/applicationTypes';
import { RESTFUL } from '@/constants/applicationServiceTypes';
import styles from './createtask.less';

import { getLatestTasks } from '../service';

const FormItem = Form.Item;
const { Option } = Select;
const formLayout = {
  labelCol: {
    span: 7,
  },
  wrapperCol: {
    span: 13,
  },
};

const cpuRequestOption = {
  fat: [
    { name: '0.1核', value: '0.1' },
    { name: '0.5核', value: '0.5' },
  ],
  test: [
    { name: '0.1核', value: '0.1' },
    { name: '0.5核', value: '0.5' },
  ],
  prod: [
    { name: '0.1核', value: '0.1' },
    { name: '0.25核', value: '0.25' },
    { name: '0.5核', value: '0.5' },
    { name: '单核', value: '1' },
    { name: '双核', value: '2' },
  ],
};

const memRequestOption = {
  fat: [
    { name: '0.25G', value: '0.25Gi' },
    { name: '0.5G', value: '0.5Gi' },
  ],
  test: [
    { name: '0.25G', value: '0.25Gi' },
    { name: '0.5G', value: '0.5Gi' },
  ],
  prd: [
    { name: '0.25G', value: '0.25Gi' },
    { name: '0.5G', value: '0.5Gi' },
    { name: '1G', value: '1Gi' },
    { name: '2G', value: '2Gi' },
    { name: '4G', value: '4Gi' },
  ],
};

const terminationGracePeriodSecOptions = [
  { name: '30秒', value: 30 },
  { name: '60秒', value: 60 },
  { name: '90秒', value: 90 },
  { name: '120秒', value: 120 },
];

/**
 *  获取资源选项
 * @param {*} resourceSpec 后端传递的resource_spec
 * @param {*} envname 环境变量为空则取：stg
 * @param {*} formName  表单中的字段含义选项只有：
 */
const getResourceSpecOptions = (resourceSpec, env, formName) => {
  const envname = env || 'test';

  // 模式数据 匹配模式
  const defaultOption = {
    mem_request: memRequestOption[envname],
    cpu_request: cpuRequestOption[envname],
  };

  if (!resourceSpec) {
    if (formName) {
      return defaultOption[formName];
    }

    return [];
  }

  return map(sortBy(resourceSpec[envname][`${formName}_list`]), (value) => ({
    name: value,
    value,
  }));
};

/**
 *  创建task
 * @param {*} props
 */

const CreateTasks = (props) => {
  const [form] = Form.useForm();

  const [showConfigCommit, setShowConfigCommit] = useState(false);
  const [isShowMorePort, setIsShowMorePort] = useState(false);
  const [isManualInput, setIsManualInput] = useState(false); // 是否手动输入镜像
  const [isAutoScale, setIsAutoScale] = useState(false); // 是否开启自动伸缩
  const [showNodeLabels, setShowNodeLabels] = useState(true); // 是否开启高级选项
  const [confirmLoading, setConfirmLoading] = useState(false);

  const {
    onSubmit,
    onCancel: handleUpdateModalVisible,
    modalVisible,
    application,
    applicationInfo,
    imageTags, // 所有镜像
    projectInfo,
    filterInfo: { envname },
    applicationTips,
    nodeLabels,
    projectClusters,
    clusterName,
    defaultClusterName,
  } = props;

  const resourceSpec =
    projectInfo && projectInfo.resource_spec ? projectInfo.resource_spec : null;

  useEffect(() => {
    if (modalVisible && application && application.id && envname) {
      if (clusterName) {
        form.setFieldsValue({
          cluster_name: clusterName,
        });
      }
      setIsAutoScale(application.type === SERVICE); // Service默认开启自动伸缩并显示最大实例数
      getLatestTasks(application.id, envname, clusterName)
        .then((result) => {
          if (!result || !result.param) {
            return;
          }

          setShowNodeLabels(false)
          const latestSetting = result.param || {};

          // 格式转换 环境变量
          const varsArray = [];
          if (latestSetting.vars) {
            const { vars } = latestSetting;
            forIn(vars, (value, key) => {
              varsArray.push({ key, value });
            });
          }

          // 格式转换 额外端口
          const exposedPortsArray = [];
          if (latestSetting.exposed_ports) {
            const exposedPorts = latestSetting.exposed_ports;
            forIn(exposedPorts, (value, key) => {
              exposedPortsArray.push({ key, value });
            });
          }

          // 判断是否有多端口暴露，存在则勾选显示更多端口
          if (exposedPortsArray.length > 0) {
            setIsShowMorePort(true);
          }

          // 判断config_commit_id存在则设置使用最新
          let setting = 0;
          if (latestSetting.config_commit_id) {
            setting = 1;
          }

          const nodeSelectorArray = [];
          if (latestSetting.node_selector) {
            const nodeSelects = latestSetting.node_selector;
            forIn(nodeSelects, (value, key) => {
              nodeSelectorArray.push({ key, value });
            });
          }

          form.setFieldsValue({
            health_check_url: latestSetting.health_check_url || '/a',
            target_port: latestSetting.target_port || 80,
            max_pod_count: latestSetting.max_pod_count || 1,
            mem_request: latestSetting.mem_request,
            cpu_limit: latestSetting.cpu_limit,
            cpu_request: latestSetting.cpu_request,
            cron_command: latestSetting.cron_command || '',
            job_command: latestSetting.job_command || '',
            backoff_limit: latestSetting.backoff_limit || 0,
            cron_param: latestSetting.cron_param,
            active_deadline_seconds:
              latestSetting?.active_deadline_seconds ?? 0,
            is_auto_scale: latestSetting.is_auto_scale || false,
            is_support_metrics: latestSetting.is_support_metrics || false,
            mem_limit: latestSetting.mem_limit,
            min_pod_count: latestSetting.min_pod_count,
            cover_command: latestSetting.cover_command || null,
            pre_stop_command: latestSetting.pre_stop_command || null,
            importance:
              latestSetting?.node_affinity_label_config?.importance ?? '',
            cpu: latestSetting?.node_affinity_label_config?.cpu ?? '',
            mem: latestSetting?.node_affinity_label_config?.mem ?? '',
            exclusive:
              latestSetting?.node_affinity_label_config?.exclusive ?? '',
            disable_high_availability:
              latestSetting?.disable_high_availability ?? true,
            disable_canary: 
              latestSetting?.disable_canary ?? false,
            termination_grace_period_sec:
              latestSetting?.termination_grace_period_sec ?? 0,
            config_mount_path: latestSetting?.config_mount_path ?? '',
            vars: varsArray,
            exposed_ports: exposedPortsArray,
            node_selector: nodeSelectorArray,
            setting,
            is_support_sticky_session: latestSetting?.is_support_sticky_session,
            // session_cookie_max_age为0时不可以覆盖初始值
            session_cookie_max_age:
              latestSetting?.session_cookie_max_age || undefined,
            // is_show_more_port,
            open_cold_storage: latestSetting?.open_cold_storage || false,
            metrics_port: latestSetting?.metrics_port || undefined,
            config_rename_prefix: latestSetting?.config_rename_prefix || '',
            config_rename_mode: latestSetting?.config_rename_mode || undefined,
          });
          setIsAutoScale(latestSetting.is_auto_scale || false);
        })
        .catch(() => {});
    }
  }, [modalVisible]);

  const create = async () => {
    try {
      setConfirmLoading(true);
      // 验证
      let fieldsValue = await form.validateFields();
      // 镜像处理
      if (isArray(fieldsValue.image_version)) {
        fieldsValue = {
          ...fieldsValue,
          image_version: fieldsValue.image_version[1],
        };
      }

      if (fieldsValue.setting === 1) {
        fieldsValue.config_commit_id = '@latest';
      }

      if (application.type === SERVICE) {
        fieldsValue.target_port = Number(fieldsValue.target_port);
      }

      // 环境变量数组转json对象
      const newVars = {};
      forEach(fieldsValue.vars, ({ value, key }) => {
        if (key && value) {
          newVars[key] = value;
        }
      });
      fieldsValue.vars = newVars;

      // 节点选择器array 转 json对象
      if (
        fieldsValue.node_selector &&
        JSON.stringify(fieldsValue.node_selector) !== '[]'
      ) {
        const newNodeSelector = {};
        forEach(fieldsValue.node_selector, ({ key, value }) => {
          if (key && value) {
            newNodeSelector[key] = value;
          }
        });
        fieldsValue.node_selector = newNodeSelector;
      } else {
        fieldsValue.node_selector = {};
      }

      if (
        fieldsValue.exposed_ports &&
        JSON.stringify(fieldsValue.exposed_ports) !== '[]'
      ) {
        const newExposedPorts = {};
        forEach(fieldsValue.exposed_ports, ({ value, key }) => {
          if (key && value) {
            newExposedPorts[key] = Number(value);
          }
        });
        fieldsValue.exposed_ports = newExposedPorts;
      }

      // 需要写死的数据
      fieldsValue.cpu_limit = fieldsValue.cpu_limit || '2';
      fieldsValue.mem_limit = fieldsValue.mem_limit || '4Gi';

      if ([CRON_JOB, ONE_TIME_JOB].includes(application.type)) {
        // message.success('我是CronJob');
        fieldsValue.concurrency_policy = 'Forbid';
        if (application.type === CRON_JOB)
          fieldsValue.restart_policy = 'OnFailure';
        fieldsValue.successful_history_limit = 5;
        fieldsValue.failed_history_limit = 3;
        fieldsValue.active_deadline_seconds = Number(
          fieldsValue.active_deadline_seconds
        );
        fieldsValue.backoff_limit = +fieldsValue.backoff_limit;
      }

      fieldsValue.node_affinity_label_config = {
        importance: fieldsValue.importance || '',
        cpu: fieldsValue.cpu || '',
        mem: fieldsValue.mem || '',
        exclusive: fieldsValue.exclusive || '',
      };
      delete fieldsValue.importance;
      delete fieldsValue.cpu;
      delete fieldsValue.mem;
      delete fieldsValue.exclusive;
      delete fieldsValue.setting;
      await onSubmit(fieldsValue);
    } finally {
      setConfirmLoading(false);
    }
  };

  const getImageTagsOption = () => {
    let branchs = [];
    const tags = [];
    forEach(imageTags, (item) => {
      branchs.push(item.branch_name);
    });

    branchs = [...new Set(branchs)];

    forEach(branchs, (branch) => {
      const childs = [];
      forEach(imageTags, (item) => {
        if (item.branch_name === branch) {
          childs.push({
            value: item.version,
            label: item.template
              ? `${item.commit_id}｜${item.template.name}｜${item.update_time}`
              : `${item.commit_id}｜${item.update_time}`,
          });
        }
      });

      tags.push({ value: branch, label: branch, children: childs });
    });
    return tags;
  };

  const settingChange = (values) => {
    if (values && values === 2) {
      setShowConfigCommit(true);
    } else {
      setShowConfigCommit(false);
    }
  };

  const changeIsShowMorePort = (e) => {
    setIsShowMorePort(e.target.checked);
  };

  // form值更新
  const handleValuesChange = (changedValues) => {
    if (changedValues.hasOwnProperty('is_auto_scale')) {
      setIsAutoScale(changedValues.is_auto_scale);
    }
  };

  const initialValues = useMemo(
    () => ({
      setting: 0,
      health_check_url: '/health',
      target_port: '80',
      min_pod_count: application.type === SERVICE && envname === 'prod' ? 2 : 1,
      max_pod_count: 3,
      // is_support_metrics:false,
      is_auto_scale: application.type === SERVICE,
      active_deadline_seconds: 0,
      vars: [],
      pre_stop_command: '',
      cover_command: '',
      importance: '', // 服务分级
      cpu: '', // 实例CPU规格
      mem: '', // 实例内存规格
      exclusive: '', // 专用标记
      disable_high_availability: true,
      termination_grace_period_sec: 30,
      config_mount_path: '',
      open_cold_storage: false,
      cluster_name: defaultClusterName,
      readiness_probe_initial_delay_seconds: 10,
      liveness_probe_initial_delay_seconds: 10,
      disable_canary: false,
    }),
    [defaultClusterName, application]
  );

  return (
    <Modal
      width={800}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title="创建任务"
      visible={modalVisible}
      okText="创建"
      onOk={create}
      onCancel={() => handleUpdateModalVisible()}
      confirmLoading={confirmLoading}
    >
      <Form
        {...formLayout}
        form={form}
        onValuesChange={(changedValues, allValues) =>
          handleValuesChange(changedValues, allValues)
        }
        initialValues={initialValues}
      >
        {(applicationTips?.wasted_max_cpu_usage_rate || applicationTips?.wasted_max_mem_usage_rate) &&
          applicationTips?.recommend_cpu_request &&
          applicationTips?.recommend_max_pod_count &&
          applicationTips?.recommend_mem_request &&
          applicationTips?.recommend_min_pod_count && (
            <Alert
              type="error"
              closable
              message={
                <div style={{ display: 'flex', flexWrap: 'wrap' }}>
                  <div style={{ marginRight: '10px' }}>
                    推荐cpu规格：
                    <strong>{applicationTips?.recommend_cpu_request}</strong>
                  </div>
                  <div style={{ marginRight: '10px' }}>
                    推荐内存规格：
                    <strong>{applicationTips?.recommend_mem_request}</strong>
                  </div>
                  <div style={{ marginRight: '10px' }}>
                    推荐最小实例数：
                    <strong>{applicationTips?.recommend_min_pod_count}</strong>
                  </div>
                  <div style={{ marginRight: '10px' }}>
                    推荐最大实例数：
                    <strong>{applicationTips?.recommend_max_pod_count}</strong>
                  </div>
                </div>
              }
            />
          )}
        <br />
        <FormItem label="应用名称">{applicationInfo?.name}</FormItem>
        <FormItem label="环境">{envname}</FormItem>

        <Form.Item label="镜像" required style={{ marginBottom: 0 }}>
          {isManualInput ? (
            <Form.Item
              style={{ display: 'inline-block', width: 'calc(100% - 80px)' }}
              name="image_version"
              rules={[{ required: true, message: '镜像信息不能为空' }]}
            >
              <Input placeholder="请输入镜像地址" />
            </Form.Item>
          ) : (
            <FormItem
              style={{ display: 'inline-block', width: 'calc(100% - 80px)' }}
              name="image_version"
              rules={[{ required: true, message: '镜像信息不能为空' }]}
            >
              <Cascader
                options={getImageTagsOption()}
                placeholder="Please select"
              />
            </FormItem>
          )}
          <Form.Item style={{ display: 'inline-block', width: '80px' }}>
            <Checkbox
              style={{ margin: '0 5px 0 10px' }}
              checked={isManualInput}
              onChange={(e) => {
                setIsManualInput(e.target.checked);
                const imageVersion = form.getFieldValue('image_version');
                // 镜像--输入框字符串格式转换为二级下拉的数组格式
                if (
                  !e.target.value &&
                  imageVersion &&
                  typeof imageVersion === 'string'
                ) {
                  const formatImageVersion = imageVersion
                    .split(':')[1]
                    .split('-')
                    .reverse();
                  form.setFieldsValue({ image_version: formatImageVersion });
                }
              }}
            />
            手动&nbsp;
            <Tooltip title="当勾选“手动”时，可直接手动输入镜像地址进行发布">
              <QuestionCircleOutlined />
            </Tooltip>
          </Form.Item>
        </Form.Item>

        {/* 高级选项 */}
        <Form.Item label="高级选项">
          <Button
            onClick={() => setShowNodeLabels(!showNodeLabels)}
            type="primary"
            size="small"
          >
            {showNodeLabels ? '收起' : '修改'}
          </Button>
        </Form.Item>

        <div
          style={showNodeLabels ? { display: 'block' } : { display: 'none' }}
        >
        <FormItem name="cluster_name" label="集群名称">
          <Select>
            {map(projectClusters, ({ name }) => (
              <Option key={name} value={name}>
                {name}
              </Option>
            ))}
          </Select>
        </FormItem>

        {projectInfo?.config_rename_prefixes?.length > 0 && (
          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) =>
              prevValues.setting !== currentValues.setting
            }
          >
            {({ getFieldValue }) =>
              !!getFieldValue('setting') && (
                <Form.Item name="config_rename_mode" label="特殊配置重命名模式">
                  <Select allowClear>
                    {map(projectInfo.config_rename_modes, (mode) => (
                      <Option key={mode.enum} value={mode.enum}>
                        {mode.name}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              )
            }
          </Form.Item>
        )}

        {/* 高级配置-收缩功能 */}

        {[CRON_JOB, ONE_TIME_JOB].includes(application.type) ? (
          <>
            <FormItem
              name={
                application.type === CRON_JOB ? 'cron_command' : 'job_command'
              }
              label="执行命令"
            >
              <Input placeholder="请输入" />
            </FormItem>
            <FormItem
              name="active_deadline_seconds"
              label="执行任务超时时间"
              rules={[{ required: true, message: '执行任务超时时间不能为空' }]}
            >
              <Input placeholder="请输入执行超时秒数(整型)" />
            </FormItem>
            {application.type === CRON_JOB ? (
              <FormItem
                name="cron_param"
                label="调度周期"
                rules={[{ required: true, message: '调度周期不能为空' }]}
              >
                <Input placeholder="请输入" />
              </FormItem>
            ) : null}
          </>
        ) : (
          <></>
        )}
        {[CRON_JOB, ONE_TIME_JOB].includes(application.type) ? (
          <Form.Item
            name="backoff_limit"
            label="失败重试次数"
            rules={[
              { required: true, message: '请输入次数' },
              {
                validator: (rule, value) => {
                  if (value && !/^-{0,1}\d+$/.test(value)) {
                    return Promise.reject(new Error('格式错误，纯数字'));
                  }
                  if (value && /^-{0,1}\d+$/.test(value) && +value < 0) {
                    return Promise.reject(new Error('值必须大于等于0'));
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input placeholder="请输入" />
          </Form.Item>
        ) : null}
        {application.type === SERVICE ? (
          <>
            <FormItem name="health_check_url" label="健康检查">
              <Input placeholder="请输入" />
            </FormItem>

            <FormItem label="端口设置" className={styles.portSetting}>
              <FormItem
                name="target_port"
                // label="端口"
              >
                <Input className={styles.targetPort} placeholder="端口" />
              </FormItem>

              <FormItem
                // name="is_show_more_port"
                className={styles.isShowMorePort}
                label="额外端口"
                // valuePropName="checked"
              >
                <Checkbox
                  checked={isShowMorePort}
                  defaultChecked={isShowMorePort}
                  onChange={changeIsShowMorePort}
                />
              </FormItem>
            </FormItem>

            {isShowMorePort ? (
              <Form.List name="exposed_ports" label="额外端口">
                {(fields, { add, remove }) => {
                  return (
                    <div>
                      {fields.map((field) => (
                        <Space
                          key={field.key}
                          style={{ display: 'flex', margin: '0px 100px' }}
                          align="start"
                        >
                          <Form.Item
                            {...field}
                            name={[field.name, 'key']}
                            fieldKey={[field.fieldKey, 'key']}
                            rules={[{ required: true, message: '缺少env key' }]}
                          >
                            <Input
                              placeholder="输入端口key"
                              className={styles.envInput}
                            />
                          </Form.Item>

                          <Form.Item
                            {...field}
                            name={[field.name, 'value']}
                            fieldKey={[field.fieldKey, 'value']}
                            rules={[
                              {
                                required: true,
                                message: '缺少env value',
                              },
                              () => ({
                                validator(rule, value) {
                                  const reG = /^[1-9]\d*$/;

                                  if (!reG.test(value)) {
                                    return Promise.reject(
                                      new Error('端口必须为数字')
                                    );
                                  }
                                  return Promise.resolve();
                                },
                              }),
                            ]}
                          >
                            <Input
                              placeholder="输入新端口"
                              className={styles.envInput}
                            />
                          </Form.Item>

                          <MinusCircleOutlined
                            onClick={() => {
                              remove(field.name);
                            }}
                          />
                        </Space>
                      ))}

                      <Form.Item>
                        <Button
                          style={{ margin: '0px 100px' }}
                          type="dashed"
                          onClick={() => {
                            add();
                          }}
                          block
                        >
                          <PlusOutlined /> 添加端口
                        </Button>
                      </Form.Item>
                    </div>
                  );
                }}
              </Form.List>
            ) : null}
          </>
        ) : (
          <></>
        )}

        <FormItem
          label="容器CPU规格"
          name="cpu_request"
          rules={[{ required: true, message: '容器CPU规格不能为空' }]}
        >
          <Select className={styles.cpuRequest}>
            {(
              getResourceSpecOptions(resourceSpec, envname, 'cpu_request') || []
            ).map((item) => {
              return (
                <Option key={item.value} value={item.value}>
                  {item.name}
                </Option>
              );
            })}
          </Select>
        </FormItem>

        <FormItem
          name="mem_request"
          label="容器内存规格"
          rules={[{ required: true, message: '容器内存规格不能为空' }]}
        >
          <Select>
            {(
              getResourceSpecOptions(resourceSpec, envname, 'mem_request') || []
            ).map((item) => {
              return (
                <Option key={item.value} value={item.value}>
                  {item.name}
                </Option>
              );
            })}
          </Select>
        </FormItem>

        {/* <Form.Item
              label="节点选择器"
            >
          </Form.Item>

            <Form.List
              name="node_selector"
              label="节点选择器"
            >
              {(fields, { add, remove }) => {
                return (
                  <div>
                    {fields.map(field => (
                      <Space key={field.key} style={{ display: 'flex', margin: '0px 100px' }} align="start">
                        <Form.Item
                          {...field}
                          name={[field.name, 'key']}
                          fieldKey={[field.fieldKey, 'key']}
                          rules={[{ required: true, message: '缺少key' }]}
                        >
                          <Input placeholder="输入key"  className={styles.envInput} />
                        </Form.Item>

                        <Form.Item
                          {...field}
                          name={[field.name, 'value']}
                          fieldKey={[field.fieldKey, 'value']}
                          rules={[{ required: true, message: '缺少value' }]}
                        >
                          <Input placeholder="输入value"  className={styles.envInput}  />
                        </Form.Item>

                        <MinusCircleOutlined
                          onClick={() => {
                            remove(field.name);
                          }}
                        />
                      </Space>
                    ))}

                    <Form.Item>
                      <Button
                          style={{ margin: '0px 100px' }}
                          type="dashed"
                          onClick={() => {
                            add();
                          }}
                          block
                        >
                          <PlusOutlined /> 添加节点选择器
                      </Button>
                    </Form.Item>
                  </div>
                );
              }}
            </Form.List>
            </>
          :
          null} */}

        {[CRON_JOB, ONE_TIME_JOB].includes(application.type) ? (
          <></>
        ) : (
          <>
            <FormItem name="min_pod_count" label="实例数">
              <InputNumber min={1} max={100000} />
            </FormItem>
            <Form.Item
              name="is_auto_scale"
              label="自动伸缩"
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
            {isAutoScale ? ( // 自动伸缩时展示最大实例数
              <FormItem
                name="max_pod_count"
                label="最大实例数"
                rules={[
                  () => ({
                    validator(rule, value) {
                      const minPodCount = form.getFieldValue('min_pod_count');
                      if (value < minPodCount) {
                        return Promise.reject(
                          new Error('最小实例数不能大于最大实例数!')
                        );
                      }
                      return Promise.resolve();
                    },
                  }),
                ]}
              >
                <InputNumber min={1} max={100000} />
              </FormItem>
            ) : null}
          </>
        )}

        <Form.Item label="环境变量">
          <Form.List name="vars">
            {(fields, { add, remove }) => {
              return (
                <div>
                  {fields.map((field) => (
                    <Space key={field.key} align="start">
                      <Form.Item
                        {...field}
                        name={[field.name, 'key']}
                        fieldKey={[field.fieldKey, 'key']}
                        rules={[{ required: true, message: '缺少env key' }]}
                      >
                        <Input
                          placeholder="输入env key"
                          className={styles.envInput}
                        />
                      </Form.Item>

                      <Form.Item
                        {...field}
                        name={[field.name, 'value']}
                        fieldKey={[field.fieldKey, 'value']}
                        rules={[{ required: true, message: '缺少env value' }]}
                      >
                        <Input
                          placeholder="输入env value"
                          className={styles.envInput}
                        />
                      </Form.Item>

                      <MinusCircleOutlined
                        onClick={() => {
                          remove(field.name);
                        }}
                      />
                    </Space>
                  ))}

                  <Form.Item>
                    <Button
                      type="dashed"
                      onClick={() => {
                        add();
                      }}
                      block
                    >
                      <PlusOutlined /> 添加环境
                    </Button>
                  </Form.Item>
                </div>
              );
            }}
          </Form.List>
        </Form.Item>

        {[CRON_JOB, ONE_TIME_JOB].includes(application.type) ? (
          <></>
        ) : (
          <>
            <Form.Item
              name="is_support_metrics"
              valuePropName="checked"
              label="采集Metrics"
            >
              <Switch />
            </Form.Item>
            <Form.Item
              noStyle
              shouldUpdate={(current, prev) =>
                current.is_support_metrics !== prev.is_support_metrics
              }
            >
              {({ getFieldValue }) => {
                const support = getFieldValue('is_support_metrics');
                return (
                  support && (
                    <Form.Item
                      name="metrics_port"
                      label="采集端口"
                      initialValue={8088}
                    >
                      <InputNumber />
                    </Form.Item>
                  )
                );
              }}
            </Form.Item>
            <FormItem name="cover_command" label="启动命令">
              <Input placeholder="请输入启动命令" />
            </FormItem>
            <FormItem name="pre_stop_command" label="预执行命令">
              <Input placeholder="请输入停止前预执行的命令" />
            </FormItem>
          </>
        )}

          <FormItem name="cpu_limit" label="容器CPU限制">
            <Select>
              {(
                getResourceSpecOptions(resourceSpec, envname, 'cpu_limit') || []
              ).map((item) => {
                return (
                  <Option key={item.value} value={item.value}>
                    {item.name}
                  </Option>
                );
              })}
            </Select>
          </FormItem>
          <FormItem name="mem_limit" label="容器内存限制">
            <Select>
              {(
                getResourceSpecOptions(resourceSpec, envname, 'mem_limit') || []
              ).map((item) => {
                return (
                  <Option key={item.value} value={item.value}>
                    {item.name}
                  </Option>
                );
              })}
            </Select>
          </FormItem>
          <Form.Item 
            name="importance"
            label={
              <>
                服务分级&nbsp;
                <Tooltip title="规格越高，可供调度的节点越多。一般情况下使用Medium即可，如有疑问可咨询运维人员">
                  <QuestionCircleOutlined />
                </Tooltip>
              </>
            }
          >
            <Select placeholder="请选择服务分级规格" allowClear>
              {(
                nodeLabels.find((i) => i.type === 'importance')?.values ?? []
              ).map((item) => (
                <Select.Option values={item} key={item}>
                  {item}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          {/* <Form.Item
            label={
              <>
                节点cpu规格&nbsp;
                <Tooltip title="节点CPU规格：服务本身运行在容器中，容器实际运行在一个节点（物理机或虚拟机）上。如果服务对运行节点的CPU有特殊要求（比如CPU主频、CPU个数等等），则应指定该规格。">
                  <QuestionCircleOutlined />
                </Tooltip>
              </>
            }
            name="cpu"
          >
            <Select placeholder="请选择节点cpu规格" allowClear>
              {(nodeLabels.find((i) => i.type === 'cpu')?.values ?? []).map(
                (item) => (
                  <Select.Option values={item} key={item}>
                    {item}
                  </Select.Option>
                )
              )}
            </Select>
          </Form.Item> */}
          {/* <Form.Item
            label={
              <>
                节点内存规格&nbsp;
                <Tooltip title="节点内存规格：服务本身运行在容器中，容器实际运行在一个节点（物理机或虚拟机）上。如果服务对运行节点的内存有特殊要求（比如内存容量），则应指定该规格。">
                  <QuestionCircleOutlined />
                </Tooltip>
              </>
            }
            name="mem"
          >
            <Select placeholder="请选择节点内存规格" allowClear>
              {(nodeLabels.find((i) => i.type === 'mem')?.values ?? []).map(
                (item) => (
                  <Select.Option values={item} key={item}>
                    {item}
                  </Select.Option>
                )
              )}
            </Select>
          </Form.Item> */}
          {[SERVICE, WORKER, ONE_TIME_JOB].includes(application.type) ? (
            <Form.Item
              label={
                <>
                  专用标记&nbsp;
                  <Tooltip title="专用标签：如果服务需要独占节点，请先联系运维人员创建专用节点，然后在该项中选择专用的标签，并且将 服务分级规格 选择为special。">
                    <QuestionCircleOutlined />
                  </Tooltip>
                </>
              }
              name="exclusive"
            >
              <Select placeholder="请选择" allowClear>
                {[SERVICE, WORKER].includes(application.type) &&
                  (
                    nodeLabels.find((i) => i.type === 'exclusive-deployment')
                      ?.values ?? []
                  ).map((item) => (
                    <Select.Option key={item} value={item}>
                      {item}
                    </Select.Option>
                  ))}
                {[ONE_TIME_JOB].includes(application.type) &&
                  (
                    nodeLabels.find((i) => i.type === 'exclusive-job')
                      ?.values ?? []
                  ).map((item) => (
                    <Select.Option key={item} value={item}>
                      {item}
                    </Select.Option>
                  ))}
              </Select>
            </Form.Item>
          ) : null}
          <Form.Item
            name="disable_high_availability"
            label="关闭实例高可用"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
          {envname === 'prod' && (
            <Form.Item
              name="disable_canary"
              label="关闭金丝雀发布"
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
          )}

          <Form.Item name="termination_grace_period_sec" label="优雅终止时长">
            <Select placeholder="请选择">
              {terminationGracePeriodSecOptions.map((item) => (
                <Select.Option
                  value={item.value}
                  key={`termination_grace_period_sec_${item.value}`}
                >
                  {item.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          {application.type === SERVICE &&
            application.service_type === RESTFUL && (
              <>
                <Form.Item
                  label="开启会话保持"
                  name="is_support_sticky_session"
                  valuePropName="checked"
                >
                  <Switch />
                </Form.Item>
                <Form.Item
                  noStyle
                  shouldUpdate={(prevValues, curValues) =>
                    prevValues.is_support_sticky_session !==
                    curValues.is_support_sticky_session
                  }
                >
                  {() =>
                    form.getFieldValue('is_support_sticky_session') ? (
                      <Form.Item required label="会话保持时间">
                        <Form.Item
                          noStyle
                          name="session_cookie_max_age"
                          initialValue={1800}
                          rules={[
                            { required: true, message: '会话保持时间不能为空' },
                            {
                              pattern: /^\d+$/,
                              message: '会话保持时间必须为正整数',
                            },
                          ]}
                        >
                          <InputNumber min={1} />
                        </Form.Item>
                        &nbsp;&nbsp;秒
                      </Form.Item>
                    ) : null
                  }
                </Form.Item>
              </>
            )}
          <Form.Item
            name="open_cold_storage"
            label="开启冷存日志"
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>
          <Form.Item
            name="readiness_probe_initial_delay_seconds"
            label="可读探针初始化时间"
          >
            <InputNumber min={10} />
          </Form.Item>
          <Form.Item
            name="liveness_probe_initial_delay_seconds"
            label="存活探针初始化时间"
          >
            <InputNumber min={10} />
          </Form.Item>
        </div>
      </Form>
    </Modal>
  );
};

export default CreateTasks;
