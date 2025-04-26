import React, { useCallback, useEffect, useMemo } from 'react';
import { Form, Button, Input, Modal, Checkbox, Space } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';
import PropTypes from 'prop-types';
import { map, omit, reduce } from 'lodash';
import Styles from './EditForm.less';

const FormItem = Form.Item;
const { TextArea } = Input;
const formLayout = {
  labelCol: {
    span: 7,
  },
  wrapperCol: {
    span: 13,
  },
};

const EditFormModal = ({ onSubmit, onCancel, visible, values }) => {
  const [form] = Form.useForm();

  const initialValues = useMemo(
    () =>
      values
        ? {
            ...values,
            extra_hooks: !!values.extra_ding_hooks,
            extra_ding_hooks:
              values.extra_ding_hooks &&
              map(values.extra_ding_hooks, (value, key) => ({ key, value })),
          }
        : null,
    [values]
  );

  const onClickFinish = useCallback(async () => {
    const fieldsValue = await form.validateFields();
    onSubmit(
      omit(
        {
          ...values,
          ...fieldsValue,
          extra_ding_hooks: reduce(
            fieldsValue.extra_ding_hooks,
            (result, hook) => ({
              ...result,
              [hook.key]: hook.value,
            }),
            {}
          ),
        },
        'extra_hooks'
      )
    );
  }, [onSubmit, form, values]);

  useEffect(() => {
    form.setFieldsValue(
      initialValues || {
        name: '',
        label: '',
        ali_alarm_name: '',
        ding_hook: '',
        extra_hooks: '',
      }
    );
  }, [initialValues]);

  const renderFooter = () => (
    <>
      <Button onClick={onCancel}>取消</Button>
      <Button type="primary" onClick={onClickFinish}>
        完成
      </Button>
    </>
  );

  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title="规则配置"
      visible={visible}
      footer={renderFooter()}
      onCancel={onCancel}
    >
      <Form {...formLayout} form={form}>
        <FormItem
          name="name"
          label="团队名称"
          rules={[
            {
              required: true,
              message: '请输入团队名称！',
            },
          ]}
        >
          <Input placeholder="请输入" />
        </FormItem>
        <FormItem
          name="label"
          label="团队标签"
          rules={[
            {
              required: true,
              message: '请输入团队名称！',
            },
          ]}
        >
          <Input placeholder="请输入" />
        </FormItem>

        <FormItem
          name="ali_alarm_name"
          label="警告联系组"
          rules={[{ required: true, message: '请输入警告联系组' }]}
        >
          <Input placeholder="请输入警告联系组" />
        </FormItem>
        <FormItem label="默认钉钉告警地址" className={Styles.hookSetting}>
          <FormItem
            name="ding_hook"
            rules={[
              {
                required: true,
                message: '请至少输入32个字符！',
                min: 32,
              },
            ]}
          >
            <TextArea rows={4} />
          </FormItem>

          <FormItem
            label="额外地址"
            className={Styles.extraHooks}
            name="extra_hooks"
            valuePropName="checked"
          >
            <Checkbox />
          </FormItem>
        </FormItem>
        <Form.Item
          noStyle
          shouldUpdate={(prevValues, currentValues) =>
            prevValues.extra_hooks !== currentValues.extra_hooks
          }
        >
          {({ getFieldValue }) =>
            getFieldValue('extra_hooks') && (
              <Form.List name="extra_ding_hooks" label="额外地址">
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
                              placeholder="输入地址key"
                              className={Styles.envInput}
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
                            ]}
                          >
                            <Input
                              placeholder="输入新地址"
                              className={Styles.envInput}
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
                          onClick={() => add()}
                          block
                          icon={<PlusOutlined />}
                          className={Styles.addHook}
                        >
                          添加地址
                        </Button>
                      </Form.Item>
                    </div>
                  );
                }}
              </Form.List>
            )
          }
        </Form.Item>
      </Form>
    </Modal>
  );
};

EditFormModal.propTypes = {
  onSubmit: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
  visible: PropTypes.bool.isRequired,
  values: PropTypes.shape({
    id: PropTypes.string,
    name: PropTypes.string,
    label: PropTypes.string,
    ali_alarm_name: PropTypes.string,
    ding_hook: PropTypes.string,
    extra_hooks: PropTypes.bool,
    extra_ding_hooks: PropTypes.objectOf(PropTypes.string),
  }),
};

EditFormModal.defaultProps = {
  values: null,
};

export default React.memo(EditFormModal);
