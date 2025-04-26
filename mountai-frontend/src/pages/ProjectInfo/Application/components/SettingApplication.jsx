import React, { useEffect } from 'react';
import { Form, Input, Modal, Switch, message, Checkbox } from 'antd';
import { RESTFUL } from '@/constants/applicationServiceTypes';

const formLayout = {
  labelCol: {
    span: 7,
  },
  wrapperCol: {
    span: 13,
  },
};

const SettingApplication = (props) => {
  const [form] = Form.useForm();

  const {
    onSubmit,
    onCancel,
    onReload,
    modalVisible,
    applicationInfo,
    filterInfo,
    // values,
  } = props;

  useEffect(() => {
    if (modalVisible) {
      form.setFieldsValue({
        enable_branch_change_notification:
          applicationInfo.enable_branch_change_notification || false,
        description: applicationInfo.description || '',
        enable_istio:
          applicationInfo.service_type === RESTFUL
            ? applicationInfo.enable_istio ?? false
            : undefined,
      });
    }
  }, [modalVisible]);

  const update = async () => {
    const { envname } = filterInfo;
    // 验证
    const fieldsValue = await form.validateFields();
    fieldsValue.id = applicationInfo.id;
    fieldsValue.env = {
      [envname]: {
        enable_branch_change_notification:
          fieldsValue.enable_branch_change_notification,
      },
    };
    try {
      await onSubmit(fieldsValue);
      message.success('保存成功');
      onReload();
    } catch (error) {
      message.error(error.message || '保存失败');
    }
    onCancel();
  };

  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title={`编辑应用${applicationInfo.name && ` - ${applicationInfo.name}`}`}
      visible={modalVisible}
      cancelText="取消"
      okText="保存"
      onOk={() => update()}
      onCancel={() => {
        onCancel();
        form.setFieldsValue({
          description: '',
          enable_branch_change_notification: false,
        });
      }}
    >
      <Form
        {...formLayout}
        form={form}
        initialValues={{
          description: '',
          enable_branch_change_notification: false,
        }}
      >
        <Form.Item
          name="enable_branch_change_notification"
          label="分支变更通知"
          valuePropName="checked"
        >
          <Switch />
        </Form.Item>
        <Form.Item name="description" label="备注">
          <Input.TextArea rows={4} placeholder="请输入备注" />
        </Form.Item>
        {/* {applicationInfo.service_type === RESTFUL && (
          <Form.Item
            label="启用 Istio"
            name="enable_istio"
            valuePropName="checked"
          >
            <Checkbox />
          </Form.Item>
        )} */}
      </Form>
    </Modal>
  );
};

export default SettingApplication;
