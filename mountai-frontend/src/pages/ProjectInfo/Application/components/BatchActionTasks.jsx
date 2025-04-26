import React, { useState, useEffect, useCallback } from 'react';
import {
  Modal,
  Form,
  Select,
  Cascader,
  Spin,
  Alert,
  TreeSelect,
  Input,
  message,
} from 'antd';
import { map, filter, includes } from 'lodash';
import { connect } from 'dva';
import {
  CRON_JOB,
  ONE_TIME_JOB,
  SERVICE,
  WORKER,
} from '@/constants/applicationTypes';
import { ACTIONS, applicationActions } from '../../../../utils/commonData';

const layout = {
  labelCol: { span: 7 },
};

const { Option } = Select;

const ALLOW_SETTING_ACTIONS = [
  ACTIONS.FULL_DEPLOY,
  ACTIONS.CANARY_DEPLOY,
  ACTIONS.RESTART,
];

const ALLOW_IMAGE_ACTIONS = [ACTIONS.FULL_DEPLOY, ACTIONS.CANARY_DEPLOY];

const isAllowSetting = (action) => includes(ALLOW_SETTING_ACTIONS, action);

const isAllowImage = (action) => includes(ALLOW_IMAGE_ACTIONS, action);

// 批量操作弹框
const BatchActionTasks = (props) => {
  const {
    visible,
    loading,
    projectId,
    applications,
    filterInfo,
    newImageTags,
    onSubmit,
    projectClusterMap,
    onCancel: oc,
    projectInfo,
  } = props;
  const [treeData, setTreeData] = useState([]); // 格式化的应用数据
  const [options, setOptions] = useState([]); // 镜像列表
  const [form] = Form.useForm();
  const getImageTagsOption = (images) => {
    const result = [];
    images.forEach((item) => {
      if (!result.find((i) => i.value === item.branch_name)) {
        const obj = {
          value: item.branch_name,
          label: item.branch_name,
          children: [
            {
              value: item.version,
              label: `${item.commit_id}｜${item.update_time}`,
            },
          ],
        };
        result.push(obj);
      } else {
        const index = result.findIndex((i) => i.value === item.branch_name);
        result[index].children.push({
          value: item.version,
          label: `${item.commit_id}｜${item.update_time}`,
        });
      }
    });
    setOptions(result);
  };

  useEffect(() => {
    getImageTagsOption(newImageTags);
  }, [newImageTags]);

  useEffect(() => {
    const result = [];
    applications?.forEach((item) => {
      if (!result.find((i) => i.title === item.type)) {
        const obj = {
          title: item.type,
          value: item.type,
          key: item.type,
          children: [],
        };
        obj.children.push({ title: item.name, value: item.id, key: item.id });
        result.push(obj);
      } else {
        const index = result.findIndex((i) => i.title === item.type);
        result[index].children.push({
          title: item.name,
          value: item.id,
          key: item.id,
        });
      }
    });
    setTreeData(result);
  }, [applications]);

  const handleSubmit = () => {
    form
      .validateFields()
      .then((values) => {
        const {
          image_version: imageVersion,
          setting,
          config_commit_id: configCommitId,
          ...param
        } = values.param ?? {};

        onSubmit({
          ...values,
          param: {
            ...param,
            image_version: imageVersion?.[1] ?? '',
            config_commit_id: setting === 1 ? '@latest' : configCommitId,
          },
        });
      })
      .catch((err) => err.message && message.error(err.message));
  };

  const onCancel = useCallback(() => {
    form.resetFields();
    oc();
  }, [oc, form]);

  return (
    <Modal
      visible={visible}
      width="550px"
      title={
        <>
          批量操作
          <span style={{ color: 'red' }}>{`（${filterInfo.envname}）`}</span>
        </>
      }
      maskClosable={false}
      onOk={handleSubmit}
      onCancel={onCancel}
    >
      <Form name="batchActionTasks" form={form} {...layout}>
        <Spin spinning={loading}>
          <Form.Item
            name="app_ids"
            label="应用"
            rules={[{ required: true, message: '请选择应用' }]}
          >
            <TreeSelect
              treeCheckable
              treeData={treeData}
              showCheckedStrategy={TreeSelect.SHOW_CHILD}
              placeholder="请选择应用（可多选）"
              style={{ width: '100%' }}
            />
          </Form.Item>
          <Form.Item name="cluster_name" label="集群">
            <Select>
              {map(
                filter(
                  projectClusterMap[projectId],
                  ({ env }) => env === filterInfo.envname
                ),
                ({ name }) => (
                  <Option key={name} value={name}>
                    {name}
                  </Option>
                )
              )}
            </Select>
          </Form.Item>
          <Form.Item
            name="action"
            label="操作"
            rules={[
              { required: true, message: '请选择操作类型' },
              {
                validator: (rules, value) => {
                  const apps = applications.filter((i) =>
                    (form.getFieldValue('app_ids') || []).includes(i.id)
                  );
                  // Service&Worker支持
                  // full_deploy (fat/stg环境) / canary_deploy (pre/prd环境) / stop / restart / delete / resume
                  // CronJob支持
                  // full_deploy / stop / delete / resume
                  // OneTimeJob支持
                  // full_deploy / delete
                  if (
                    value &&
                    value === ACTIONS.FULL_DEPLOY &&
                    !['fat', 'test'].includes(filterInfo.envname) &&
                    apps.some((i) => [SERVICE, WORKER].includes(i.type))
                  ) {
                    return Promise.reject(
                      new Error(
                        '批量全量发布仅支持Service/Worker应用fat/test环境'
                      )
                    );
                  }
                  if (
                    value &&
                    value === ACTIONS.CANARY_DEPLOY &&
                    (!applications.some((i) =>
                      [SERVICE, WORKER].includes(i.type)
                    ) ||
                      (applications.every((i) =>
                        [SERVICE, WORKER].includes(i.type)
                      ) &&
                        !['pre', 'prd'].includes(filterInfo.envname)))
                  ) {
                    return Promise.reject(
                      new Error(
                        '批量金丝雀发布仅支持Service/Worker应用pre/prd环境'
                      )
                    );
                  }
                  if (
                    value &&
                    [ACTIONS.STOP, ACTIONS.RESUME].includes(value) &&
                    apps.some((i) => [ONE_TIME_JOB].includes(i.type))
                  ) {
                    return Promise.reject(
                      new Error('批量停止/恢复操作不支持OneTimeJob应用')
                    );
                  }
                  if (
                    value &&
                    value === ACTIONS.RESTART &&
                    apps.some((i) => [CRON_JOB, ONE_TIME_JOB].includes(i.type))
                  ) {
                    return Promise.reject(
                      new Error('批量重启操作不支持CronJob/OneTimeJob应用')
                    );
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Select placeholder="请选择操作类型">
              {applicationActions.map((item) => (
                <Option value={item.value} key={`app_act_${item.value}`}>
                  {item.name}
                </Option>
              ))}
            </Select>
          </Form.Item>
        </Spin>
        <Form.Item
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.action !== currentValues.action
          }
        >
          {({ getFieldValue }) =>
            isAllowImage(getFieldValue('action')) && (
              <Form.Item
                name={['param', 'image_version']}
                label="镜像"
                rules={[{ required: true, message: '请选择镜像' }]}
              >
                <Cascader
                  placeholder="请选择操作的镜像版本"
                  options={options}
                />
              </Form.Item>
            )
          }
        </Form.Item>
        <Form.Item
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.action !== currentValues.action
          }
        >
          {({ getFieldValue }) =>
            isAllowSetting(getFieldValue('action')) && (
              <Form.Item name={['param', 'setting']} label="配置">
                <Select>
                  <Option value={0}> 不使用 </Option>
                  <Option value={1}> 使用最新 </Option>
                  <Option value={2}> 自定义 </Option>
                </Select>
              </Form.Item>
            )
          }
        </Form.Item>
        <Form.Item
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.param?.setting !== currentValues.param?.setting ||
            prevValues.action !== currentValues.action
          }
        >
          {({ getFieldValue }) =>
            isAllowSetting(getFieldValue('action')) &&
            getFieldValue(['param', 'setting']) === 2 && (
              <Form.Item
                name={['param', 'config_commit_id']}
                label="配置ID"
                rules={[{ required: true, message: '请填写配置ID' }]}
              >
                <Input placeholder="请输入配置ID" />
              </Form.Item>
            )
          }
        </Form.Item>
        {projectInfo?.config_rename_prefixes?.length > 0 && (
          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) =>
              prevValues.param?.setting !== currentValues.param?.setting ||
              prevValues.action !== currentValues.action
            }
          >
            {({ getFieldValue }) =>
              isAllowSetting(getFieldValue('action')) &&
              !!getFieldValue(['param', 'setting']) && (
                <Form.Item
                  name={['param', 'config_rename_prefix']}
                  label="特殊配置重命名前缀"
                >
                  <Select allowClear>
                    {map(projectInfo.config_rename_prefixes, ({ prefix }) => (
                      <Option key={prefix} value={prefix}>
                        {prefix}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              )
            }
          </Form.Item>
        )}

        {projectInfo?.config_rename_prefixes?.length > 0 && (
          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) =>
              prevValues.param?.setting !== currentValues.param?.setting ||
              prevValues.action !== currentValues.action
            }
          >
            {({ getFieldValue }) =>
              isAllowSetting(getFieldValue('action')) &&
              !!getFieldValue(['param', 'setting']) && (
                <Form.Item
                  name={['param', 'config_rename_mode']}
                  label="特殊配置重命名模式"
                >
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
        <Form.Item>
          <Alert
            type="warning"
            message={
              <>
                <h4>提示：</h4>
                <p>
                  发布操作默认使用上一次部署成功的任务参数，如需更改参数，请单独发布
                </p>
                <p>
                  Service & Worker支持操作：全量发布(fat/stg环境) /
                  金丝雀发布(pre/prd环境) / 停止 / 重启 / 删除 / 恢复
                </p>
                <p>CronJob支持操作：全量发布 / 停止 / 删除 / 恢复</p>
                <p>OneTimeJob支持操作：全量发布 / 删除</p>
              </>
            }
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default connect((state) => ({
  projectClusterMap: state.application.projectClusterMap,
}))(BatchActionTasks);
