import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Collapse,
  Table,
  Button,
  Modal,
  Form,
  Input,
  message,
  Space,
  Popconfirm,
} from 'antd';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';
import {
  getProjectVariables,
  addProjectVariable,
  editProjectVariable,
  deleteProjectVariable,
} from '../../service';
import { hasDeveloperPermission } from '../../utils/hasPermission';

const { Panel } = Collapse;

function CommonParams({ projectId, memberRole, filterInfo }) {
  const [variables, setVariables] = useState([]);
  const [visible, setVisible] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [modalTitle, setModalTitle] = useState('');
  const [form] = Form.useForm();
  const [refreshTimestamp, setRefreshTimestamp] = useState(Date.now());

  const add = useCallback(() => {
    setVisible(true);
    setModalTitle('新增');
    form.resetFields();
  }, [form, setVisible, setModalTitle]);

  const handleOk = useCallback(() => {
    form.submit();
  }, [form]);

  const handleCancel = useCallback(() => {
    setVisible(false);
  }, [setVisible]);

  const onFinish = useCallback(
    async (values) => {
      setConfirmLoading(true);
      try {
        if (values.id) {
          await editProjectVariable(values.id, values.key, values.value);
        } else {
          await addProjectVariable(projectId, values.key, values.value);
        }
        setVisible(false);
        setRefreshTimestamp(Date.now());
      } catch (error) {
        message.error(error.message);
      }
      setConfirmLoading(false);
    },
    [setConfirmLoading, setVisible, setRefreshTimestamp, projectId]
  );

  const edit = useCallback(
    (r) => {
      setModalTitle('编辑');
      setVisible(true);
      form.setFieldsValue(r);
    },
    [form, setModalTitle, setVisible]
  );

  const del = useCallback(
    async (r) => {
      try {
        await deleteProjectVariable(r.id);
        setRefreshTimestamp(Date.now());
      } catch (error) {
        message.error(error.message);
      }
    },
    [setRefreshTimestamp]
  );

  const columns = useMemo(
    () => [
      {
        title: 'Key',
        dataIndex: 'key',
        key: 'key',
      },
      {
        title: 'Value',
        dataIndex: 'value',
        key: 'value',
      },
      {
        title: '操作',
        key: 'action',
        render: (_, r) =>
          hasDeveloperPermission(memberRole, filterInfo.envname) ? (
            <Space size="middle">
              <Button
                icon={<EditOutlined />}
                type="primary"
                shape="circle"
                onClick={() => edit(r)}
              />
              <Popconfirm
                title={`确认删除变量${r.key}：${r.value}吗？`}
                onConfirm={() => del(r)}
                okText="确认"
                cancelText="取消"
              >
                <Button
                  icon={<DeleteOutlined />}
                  type="danger"
                  shape="circle"
                />
              </Popconfirm>
            </Space>
          ) : null,
      },
    ],
    [edit, del, memberRole, filterInfo]
  );

  useEffect(() => {
    getProjectVariables(projectId).then((result) =>
      setVariables(result?.list ?? [])
    );
  }, [refreshTimestamp]);

  return (
    <Collapse style={{ marginBottom: '1rem' }}>
      <Panel header="通用参数" key="1">
        {hasDeveloperPermission(memberRole, filterInfo.envname) && (
          <Button type="primary" onClick={add}>
            新增
          </Button>
        )}
        <Modal
          destroyOnClose
          title={modalTitle}
          visible={visible}
          onOk={handleOk}
          confirmLoading={confirmLoading}
          onCancel={handleCancel}
        >
          <Form form={form} onFinish={onFinish}>
            <Form.Item noStyle name="id" />
            <Form.Item label="Key" name="key" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
            <Form.Item label="Value" name="value" rules={[{ required: true }]}>
              <Input />
            </Form.Item>
          </Form>
        </Modal>
        <Table dataSource={variables} columns={columns} />
      </Panel>
    </Collapse>
  );
}

export default React.memo(CommonParams);
