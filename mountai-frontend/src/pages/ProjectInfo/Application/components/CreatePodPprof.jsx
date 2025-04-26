import React, { useState } from 'react';
import { Modal, Form, Select, Input, Spin, Button, message } from 'antd';
import {
  pprofTypeConfig,
  pprofActionConfig,
} from '../../../../utils/commonData';
import { createPodPprofData } from '../service';

// Go语言项目为pod创建pprof
const CreatePodPprof = ({
  visible,
  actPod,
  actRunning,
  actApplication,
  projectInfo,
  filterInfo,
  onCancel,
}) => {
  const [form] = Form.useForm();
  const [mustSec, setMustSec] = useState(false); // 是否必填second字段
  const [loading, setLoading] = useState(false); // 是否正在调接口获取

  const handleSubmit = () => {
    if (loading) return;
    form
      .validateFields()
      .then((values) => {
        setLoading(true);
        const { version } = actRunning;
        const podName = actPod.name;
        const query = {
          env_name: filterInfo.envname,
        };
        const body = {
          env_name: filterInfo.envname,
          app_id: actApplication.id,
          type: values.type,
          action: values.action,
          seconds: +values.seconds,
          pod_port: +values.pod_port,
          cluster_name: actApplication.cluster_name,
          namespace: actRunning.namespace,
        };
        createPodPprofData(version, podName, query, body)
          .then((res) => {
            if (values.action === 'svg') {
              // svg图片展示
              const svg = new Blob([res], {
                type: 'image/svg+xml;charset=utf-8',
              });
              const DOMURL =
                window.self.URL || window.self.webkitURL || window.self;
              const url = DOMURL.createObjectURL(svg);
              window.open(url);
            } else if (['tree', 'top'].includes(values.action)) {
              // 文本数据直接展示
              const wind = window.open('', '');
              const dom = document.createElement('div');
              dom.style.whiteSpace = 'pre-line';
              dom.innerHTML = res;
              wind.document.body.appendChild(dom);
            } else {
              // download下载
              const blob = new Blob([res]);
              const href = window.URL.createObjectURL(blob);
              const dom = document.createElement('a');
              dom.href = href;
              dom.download = `${projectInfo.name}-${
                actApplication.name
              }-${new Date().getTime()}.pprof`;
              document.body.appendChild(dom);
              dom.click();
              document.body.removeChild(dom);
              window.URL.revokeObjectURL(href);
            }
            setLoading(false);
          })
          .catch((err) => {
            message.error(err);
            setLoading(false);
          });
      })
      .catch((err) => message.error(err));
  };

  const formLayout = {
    labelCol: {
      span: 4,
    },
    wrapperCol: {
      span: 20,
    },
  };
  return (
    <Modal
      title={`获取pprof${actApplication.name && ` - ${actApplication.name}`}`}
      visible={visible}
      width="800px"
      footer={null}
      onCancel={() => onCancel()}
    >
      <Spin spinning={loading} tip="正在获取中...">
        <Form
          name="create_pod_pprof"
          form={form}
          {...formLayout}
          initialValues={{
            type: 'allocs',
            action: 'svg',
            seconds: 15,
            pod_port: 8089,
          }}
        >
          <Form.Item label="类型" name="type" required>
            <Select
              placeholder="请选择"
              onChange={(value) => {
                if (['profile', 'trace'].includes(value)) {
                  form.setFieldsValue({ action: 'download' });
                  setMustSec(true);
                } else {
                  setMustSec(false);
                }
              }}
            >
              {pprofTypeConfig.map((item) => (
                <Select.Option
                  key={`pprof_type_${item.value}`}
                  value={item.value}
                >
                  {item.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item label="操作" name="action" required>
            <Select placeholder="请选择">
              {(mustSec
                ? pprofActionConfig.filter((i) => i.value === 'download')
                : pprofActionConfig
              ).map((item) => (
                <Select.Option
                  key={`pprof_action_${item.value}`}
                  value={item.value}
                >
                  {item.name}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          {mustSec ? (
            <Form.Item
              label="采集秒数"
              name="seconds"
              rules={[
                { required: true, message: '请输入秒数' },
                {
                  validator: (rules, value) => {
                    if (value && !/^[1-9][0-9]*$/.test(value)) {
                      return Promise.reject(new Error('请输入大于1的纯数字'));
                    }
                    return Promise.resolve();
                  },
                },
              ]}
            >
              <Input placeholder="请输入秒数" />
            </Form.Item>
          ) : null}
          <Form.Item
            label="pprof端口号"
            name="pod_port"
            rules={[
              { required: true, message: '请输入端口号' },
              {
                validator: (rules, value) => {
                  if (value && !/^\d+$/.test(value)) {
                    return Promise.reject(new Error('请输入纯数字'));
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input placeholder="请输入端口号" />
          </Form.Item>
        </Form>
        <div
          style={{ width: '100%', display: 'flex', justifyContent: 'flex-end' }}
        >
          <Button type="default" onClick={() => onCancel()}>
            取消
          </Button>
          &nbsp;&nbsp;&nbsp;
          <Button
            type="primary"
            loading={loading}
            onClick={() => handleSubmit()}
          >
            获取
          </Button>
        </div>
      </Spin>
    </Modal>
  );
};

export default CreatePodPprof;
