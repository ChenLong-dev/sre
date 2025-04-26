import React, { useCallback, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Form, InputNumber, Modal } from 'antd';
import { getLatestTasks } from '../service';

const formLayout = {
  labelCol: {
    span: 6,
  },
  wrapperCol: {
    span: 14,
  },
};

const UpdateHpaModal = ({
  visible,
  application,
  filterInfo: { envname },
  onSubmit: os,
  onCancel,
  clusterName,
}) => {
  const [form] = Form.useForm();

  const onSubmit = useCallback(async () => {
    const values = await form.validateFields();
    os(values);
  }, [os, form]);

  useEffect(() => {
    (async () => {
      try {
        const latestTasks = await getLatestTasks(
          application.id,
          envname,
          clusterName
        );
        form.setFieldsValue({
          min_pod_count: latestTasks?.param?.min_pod_count,
          max_pod_count: latestTasks?.param?.max_pod_count,
        });
      } catch (_) {
        //
      }
    })();
  }, [application, envname, form, visible]);

  return (
    <Modal
      width={640}
      destroyOnClose
      title="配置弹性伸缩"
      visible={visible}
      cancelText="取消"
      okText="保存"
      onOk={onSubmit}
      onCancel={onCancel}
    >
      <Form {...formLayout} form={form} preserve={false}>
        <Form.Item
          name="min_pod_count"
          label="最小实例数"
          rules={[
            {
              required: true,
              message: '请输入最小实例数！',
            },
          ]}
        >
          <InputNumber min={1} />
        </Form.Item>
        <Form.Item
          name="max_pod_count"
          label="最大实例数"
          dependencies={['min_pod_count']}
          rules={[
            {
              required: true,
              message: '请输入最大实例数！',
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || value >= getFieldValue('min_pod_count')) {
                  return Promise.resolve();
                }
                return Promise.reject(
                  new Error('最大实例数应大于等于最小实例数！')
                );
              },
            }),
          ]}
        >
          <InputNumber min={1} />
        </Form.Item>
      </Form>
    </Modal>
  );
};

UpdateHpaModal.propTypes = {
  visible: PropTypes.bool.isRequired,
  application: PropTypes.shape({
    id: PropTypes.string,
  }),
  filterInfo: PropTypes.shape({
    envname: PropTypes.string,
  }),
  onSubmit: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
};

export default React.memo(UpdateHpaModal);
