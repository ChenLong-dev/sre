import { Divider } from 'antd';
import React, { useState, useRef, useEffect, useCallback } from 'react';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import StatusIcon from '@/components/StatusIcon';
import { queryGitInfo } from '@/services/git';
import showBuildLogModal from '@/utils/showBuildLogModal';
import { queryRule } from './service';

const TableList = (props) => {
  const {
    match: {
      params: { projectId },
    },
  } = props;
  const [gitInfo, setGitInfo] = useState({});

  const openImageLog = useCallback(
    (buildId) => showBuildLogModal(projectId, buildId),
    [projectId]
  );

  const actionRef = useRef();
  const columns = [
    {
      title: '分支',
      dataIndex: 'branch_name',
      hideInSearch: true,
    },
    {
      title: 'commit_id',
      dataIndex: 'commit_id',
      // hideInTable:true,
      hideInForm: true,
      render: (_, record) => (
        <>
          <a
            href={`${
              gitInfo && gitInfo.web_url ? gitInfo.web_url : ''
            }/commit/${record.commit_id}`}
            target="_blank"
            rel="noreferrer"
          >
            {record.commit_id}
          </a>
          {/* <Divider type="vertical" /> */}
        </>
      ),
    },
    {
      title: '最后提交记录',
      dataIndex: 'last_comment',
      valueType: 'text',
      hideInSearch: true,
    },
    {
      title: '构建时间',
      dataIndex: 'create_time',
      valueType: 'text',
      hideInSearch: true,
    },
    {
      title: '持续时间',
      dataIndex: 'duration',
      valueType: 'text',
      hideInSearch: true,
    },
    {
      title: '描述',
      dataIndex: 'description',
      hideInSearch: true,
    },
    {
      title: '操作人',
      dataIndex: ['user_profile', 'name'],
      hideInSearch: true,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <>
          <StatusIcon status={record.status} />
          <Divider type="vertical" />
          <a
            onClick={() => {
              openImageLog(record.build_id);
            }}
          >
            {/* openImageLog */}
            构建日志
          </a>
          <Divider type="vertical" />
          <a href={`${record.job_url}`} target="_blank" rel="noreferrer">
            构建任务
          </a>
        </>
      ),
    },
  ];

  useEffect(() => {
    async function fetchData() {
      const response = await queryGitInfo(projectId);
      return response;
    }

    fetchData().then((res) => {
      setGitInfo(res);
    });
  }, []);

  return (
    <PageHeaderWrapper title={false}>
      <ProTable
        // headerTitle="查询表格"
        actionRef={actionRef}
        onSearch={(keyword) => {
          console.log('搜索', keyword);
        }}
        // toolBarRender={(action, { selectedRows }) => [
        //   <Button type="primary" onClick={() => handleModalVisible(true)}>
        //     <PlusOutlined /> 新建
        //   </Button>,
        // ]}

        request={(params, sorter, filter) =>
          queryRule({ ...params, projectId, sorter, filter })
        }
        columns={columns}
        rowKey="build_id"
        search={false}
      />

      {/* https://git2.qingtingfm.com/web/qt-application-manage_system/commit/1dcf6edc */}
    </PageHeaderWrapper>
  );
};

export default TableList;
