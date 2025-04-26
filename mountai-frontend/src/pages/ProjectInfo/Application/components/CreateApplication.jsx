import React from 'react';
import { Form, Button, Input, Modal, Select, Switch, Checkbox } from 'antd';
import {
  applicationServiveType,
  applicationType,
  applicationSentryType,
} from '@/utils/commonData';
import { RESTFUL } from '@/constants/applicationServiceTypes';
import { SERVICE } from '@/constants/applicationTypes';

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

const CreateApplication = (props) => {
  const [form] = Form.useForm();

  const {
    onSubmit,
    onCancel: handleUpdateModalVisible,
    modalVisible,
    projectInfo,
    // values,
  } = props;

  const create = async () => {
    // 验证
    form
      .validateFields()
      .then((values) => {
        const query = { ...values };
        if (query?.sentry_project_slug_type === 2) {
          query.sentry_project_slug = '#create';
        }
        delete query?.sentry_project_slug_type;
        onSubmit(query);
      })
      .catch(() => null);
  };

  const renderFooter = () => {
    return (
      <>
        <Button
          onClick={() => {
            handleUpdateModalVisible();
            form.setFieldsValue({
              type: '',
              service_type: '',
              name: '',
              enable_istio: false,
            });
          }}
        >
          取消
        </Button>
        <Button type="primary" onClick={() => create()}>
          创建
        </Button>
      </>
    );
  };

  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title="创建应用"
      visible={modalVisible}
      footer={renderFooter()}
      onCancel={() => {
        handleUpdateModalVisible();
        form.setFieldsValue({
          type: '',
          service_type: '',
          name: '',
          enable_istio: false,
        });
      }}
    >
      <Form
        {...formLayout}
        form={form}
        initialValues={{
          name: '',
          service_type: '',
          type: '',
          description: '',
          enable_istio: false,
          enable_branch_change_notification: false,
        }}
      >
        <FormItem
          name="name"
          label="应用名称"
          rules={[
            {
              required: true,
              message: '应用名称为必填项',
            },
            {
              validator: (rules, value) => {
                if (
                  value &&
                  !/^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$/.test(value)
                ) {
                  return Promise.reject(new Error('格式错误'));
                }
                if (
                  value &&
                  value.length > 50 - (projectInfo.name.length + 16)
                ) {
                  return Promise.reject(
                    new Error(
                      `超出最大字符限制，不能超过${
                        50 - (projectInfo.name.length + 16)
                      }字`
                    )
                  );
                }
                return Promise.resolve();
              },
            },
          ]}
        >
          <Input placeholder="" />
        </FormItem>

        <FormItem
          name="type"
          label="应用类型"
          rules={[
            {
              required: true,
              message: '请选择应用类型',
            },
          ]}
        >
          <Select>
            {applicationType &&
              applicationType.map((item) => {
                return (
                  <Option key={item.value} value={item.value}>
                    {item.name}
                  </Option>
                );
              })}
          </Select>
        </FormItem>
        <FormItem
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.type !== currentValues.type
          }
        >
          {({ getFieldValue }) =>
            getFieldValue('type') === SERVICE && (
              <FormItem
                name="service_type"
                label="服务类型"
                rules={[
                  {
                    required: true,
                    message: '请选择服务类型',
                  },
                ]}
                preserve={false}
              >
                <Select>
                  {applicationServiveType &&
                    applicationServiveType.map((item) => {
                      return (
                        <Option key={item.value} value={item.value}>
                          {item.name}
                        </Option>
                      );
                    })}
                </Select>
              </FormItem>
            )
          }
        </FormItem>
        {/* <FormItem
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.service_type !== currentValues.service_type ||
            prevValues.type !== currentValues.type
          }
        >
          {({ getFieldValue }) =>
            getFieldValue('type') === SERVICE &&
            getFieldValue('service_type') === RESTFUL && (
              <FormItem
                label="启用 Istio"
                name="enable_istio"
                valuePropName="checked"
                preserve={false}
              >
                <Checkbox />
              </FormItem>
            )
          }
        </FormItem> */}
        {/* sentry类型 */}
        <Form.Item name="sentry_project_slug_type" label="sentry项目">
          <Select placeholder="请选择">
            {applicationSentryType.map((item) => (
              <Select.Option
                key={`sentry_type_${item.value}`}
                value={item.value}
              >
                {item.name}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>
        <FormItem
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.sentry_project_slug_type !==
            currentValues.sentry_project_slug_type
          }
        >
          {({ getFieldValue }) =>
            getFieldValue('sentry_project_slug_type') === 3 && (
              <Form.Item
                name="sentry_project_slug"
                label="sentry项目名"
                rules={[{ required: true, message: '请输入自定义sentry值' }]}
              >
                <Input placeholder="请输入" />
              </Form.Item>
            )
          }
        </FormItem>
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
      </Form>
    </Modal>
  );
};

export default CreateApplication;
