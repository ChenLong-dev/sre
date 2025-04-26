import { EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { Button, Modal, Space, message, Input, Popconfirm, Form } from 'antd';
import React, { useState, useRef, useCallback, useMemo } from 'react';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import { useParams } from 'umi';
import {
  getTemplates,
  addTemplate,
  editTemplate,
  deleteTemplate,
} from './service';

const TableList = () => {
  const { teamId } = useParams();
  const actionRef = useRef();
  const [visible, setVisible] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [modalTitle, setModalTitle] = useState('');
  const [form] = Form.useForm();

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
          await editTemplate(values.id, values.name, values.content);
        } else {
          await addTemplate(teamId, values.name, values.content);
        }
        setVisible(false);
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } catch (error) {
        message.error(error.message);
      }
      setConfirmLoading(false);
    },
    [setVisible, teamId]
  );

  const edit = useCallback(
    (r) => {
      setModalTitle('编辑');
      setVisible(true);
      form.setFieldsValue(r);
    },
    [form, setModalTitle, setVisible]
  );

  const del = useCallback(async (r) => {
    try {
      await deleteTemplate(r.id);
      if (actionRef.current) {
        actionRef.current.reload();
      }
    } catch (error) {
      message.error(error.message);
    }
  }, []);

  const columns = useMemo(
    () => [
      {
        title: '模板名称',
        dataIndex: 'name',
      },
      {
        title: '模板内容',
        dataIndex: 'content',
      },
      {
        title: '操作',
        dataIndex: 'action',
        render: (_, r) => (
          <Space size="middle">
            <Button
              icon={<EditOutlined />}
              type="primary"
              shape="circle"
              onClick={() => edit(r)}
            />
            <Popconfirm
              title={`确认删除模板${r.name}吗？`}
              onConfirm={() => del(r)}
              okText="确认"
              cancelText="取消"
            >
              <Button icon={<DeleteOutlined />} type="danger" shape="circle" />
            </Popconfirm>
          </Space>
        ),
      },
    ],
    [edit, del]
  );

  const request = useCallback(
    async ({ pageSize, current }) => {
      try {
        const result = await getTemplates(teamId, current, pageSize);
        return {
          current: result.page,
          data: result.list,
          pageSize: result.limit,
          total: result.count,
          success: true,
        };
      } catch (e) {
        return {
          success: false,
        };
      }
    },
    [teamId]
  );

  return (
    <PageHeaderWrapper title={false}>
      <ProTable
        actionRef={actionRef}
        toolBarRender={() => [
          <Button type="primary" onClick={add}>
            新增
          </Button>,
        ]}
        request={request}
        columns={columns}
        rowKey="id"
        search={false}
      />
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
          <Form.Item label="模板名称" name="name" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item
            label="模板内容"
            name="content"
            rules={[{ required: true }]}
          >
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </PageHeaderWrapper>
  );
};

export default TableList;
