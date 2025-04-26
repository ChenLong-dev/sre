import { PlusOutlined } from '@ant-design/icons';
import { Button, message, Space } from 'antd';
import React, { useState, useRef, useCallback } from 'react';
import { Link } from 'umi';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import EditFormModal from './components/EditFormModal';
import DeleteConfirm from './components/DeleteConfirm';

import { queryRule, udateTeam, addRule, removeRule } from './service';
/**
 * 添加节点
 * @param fields
 */

const TableList = () => {
  const [modalVisible, setModalVisible] = useState(false);

  const [isCreate, setIsCreate] = useState(false);
  const [stepFormValues, setStepFormValues] = useState({});
  const actionRef = useRef();

  const removeWork = (record) => {
    removeRule(record);
    if (actionRef.current) {
      actionRef.current.reload();
    }
  };

  const columns = [
    {
      title: '团队名称',
      dataIndex: 'name',

      rules: [
        {
          required: true,
          message: '规则名称为必填项',
        },
      ],
      hideInSearch: true,
    },
    {
      title: '关键词搜索',
      dataIndex: 'keyword',
      hideInTable: true,
      hideInForm: true,
      rules: [
        {
          // required: true,
          message: '请输入关键词',
        },
      ],
    },
    {
      title: '团队标签',
      dataIndex: 'label',
      valueType: 'text',
      rules: [{ required: true, message: '请输入团队标签' }],
      hideInSearch: true,
    },
    {
      title: '警告联系组',
      dataIndex: 'ali_alarm_name',
      valueType: 'text',
      rules: [{ required: true, message: '请输入警告联系组' }],
      hideInSearch: true,
    },
    {
      title: '飞书地址',
      dataIndex: 'ding_hook',
      valueType: 'textarea',
      hideInSearch: true,
      rules: [{ required: false, message: '请输入飞书地址' }],
      render: (text, record) => (
        <div
          style={{
            maxWidth: '250px',
            // whiteSpace: 'nowrap',
            // overflow: 'hidden',
            wordWrap: 'break-word',
            wordBreak: 'normal',
            // textOverflow: 'ellipsis'
          }}
        >
          {record.ding_hook}
        </div>
      ),
    },

    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <Space>
          <a
            onClick={() => {
              setModalVisible(true);
              setIsCreate(false);
              setStepFormValues(record);
            }}
          >
            编辑
          </a>
          <DeleteConfirm record={record} removeRule={removeWork} />
          {/* <Link to={`/team/${record.id}/image_arg_template`}>镜像参数模板</Link> */}
        </Space>
      ),
    },
  ];

  const onSubmit = useCallback(
    async (fields) => {
      const hide = message.loading(isCreate ? '正在添加' : '正在配置');
      try {
        if (isCreate) {
          await addRule({ ...fields });
        } else {
          await udateTeam({ ...fields });
        }
        hide();
        message.success(isCreate ? '添加成功' : '配置成功');
        setModalVisible(false);
        setStepFormValues({});
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } catch (error) {
        message.error(
          error.message || (isCreate ? '添加失败请重试！' : '配置失败请重试！')
        );
      }
    },
    [isCreate]
  );

  return (
    <PageHeaderWrapper title={false}>
      <ProTable
        actionRef={actionRef}
        toolBarRender={() => [
          <Button
            type="primary"
            onClick={() => {
              setModalVisible(true);
              setIsCreate(true);
            }}
          >
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={(params, sorter, filter) =>
          queryRule({ ...params, sorter, filter })
        }
        columns={columns}
        rowKey="id"
      />
      <EditFormModal
        onSubmit={onSubmit}
        onCancel={() => {
          setModalVisible(false);
          setStepFormValues({});
        }}
        visible={modalVisible}
        values={isCreate ? null : stepFormValues}
      />
    </PageHeaderWrapper>
  );
};

export default TableList;
