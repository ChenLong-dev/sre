import { Button, message, Modal } from 'antd';
import React, { useState, useRef } from 'react';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';

import { applicationType, isFav } from '@/utils/commonData';
import { isEmpty, isEqual, pick, reduce } from 'lodash';
import styles from './index.less';
import { queryActivities } from './service';

// 应用类型  应用服务类型

const arrayToProTableColumns = (datas) =>
  reduce(
    datas,
    (result, { value, name }) => ({
      ...result,
      [value]: { text: name },
    }),
    {}
  );

const TableList = (props) => {
  const {
    location: { query },
  } = props;

  const [newLoad, setNewLoad] = useState(true);
  const actionRef = useRef();

  const getActivities = ({ ...params }) => {
    if (newLoad) {
      setNewLoad(false);
      return queryActivities({ ...params, ...query });
    }

    const newquery = pick(params, [
      'env_name',
      'app_type',
      'app_name',
      'pageSize',
      'current',
    ]);
    if (!isEqual(newquery, query)) {
      if (!isEmpty(newquery)) {
        const Url = new URLSearchParams(newquery);
        window.history.replaceState(null, null, `?${Url.toString()}`);
      } else {
        window.history.replaceState(null, null, `?`);
      }
    }
    return queryActivities({ ...newquery });
  };

  /**
   *  打开运行日志
   */
  const openRunningInfoLog = (record) => {
    if (record.detail && typeof record.detail === 'string') {
      Modal.info({
        title: '日志',
        width: 950,
        maskClosable: true,
        content: <div className={styles.preLine}>{record.detail}</div>,
        onOk() {},
      });
    } else {
      message.error('日志不存在');
    }
  };

  const columns = [
    {
      title: '环境',
      dataIndex: 'env_name',
      initialValue: query && query.env_name ? query.env_name : '',
      valueEnum: {
        test: {
          text: 'test',
          env_name: 'test',
        },
        prod: {
          text: 'prod',
          env_name: 'prod',
        },
        fat: {
          text: 'fat',
          env_name: 'fat',
        },
      },
      // hideInSearch:true
    },
    {
      title: '操作类型',
      dataIndex: 'action_display',
      hideInSearch: true,
      // hideInTable:true,
      hideInForm: true,
      render: (text, record) => (
        <div
          style={{
            maxWidth: '120px',
          }}
        >
          {record.action_display}
        </div>
      ),
    },
    {
      title: '项目名称',
      dataIndex: 'project_name',
      hideInSearch: true,
      render: (_, record) => (
        <>
          <a
            href={`/project/${record.project_id}/application `}
            // target='_blank' rel='noopener noreferer'
          >
            {record.project_name}
          </a>
        </>
      ),
    },
    {
      title: '版本',
      dataIndex: 'version',
      hideInSearch: true,
      render: (text, record) => (
        <div
          style={{
            maxWidth: '150px',
            // overflow: 'hidden',
            // textOverflow: 'ellipsis',
            // whiteSpace: 'nowrap',
            // wordWrap:'break-word',
            // wordBreak:'normal',
          }}
        >
          {record.version}
        </div>
      ),
    },
    {
      title: '应用类型',
      dataIndex: 'app_type',
      initialValue: query && query.app_type ? query.app_type : '',
      valueType: 'text',
      // dataIndex: 'teamid',
      valueEnum: arrayToProTableColumns(applicationType),
      // hideInSearch:true
    },
    {
      title: '应用名称',
      initialValue: query && query.app_name ? query.app_name : '',
      dataIndex: 'app_name',
      valueType: 'textarea',
    },
    {
      title: '操作人',
      hideInSearch: true,
      dataIndex: 'operator_name',
      valueType: 'text',
    },
    {
      title: '状态显示',
      hideInSearch: true,
      dataIndex: 'status_display',
      valueType: 'text',
    },

    {
      title: '操作时间',
      hideInSearch: true,
      dataIndex: 'create_time',
      valueType: 'dateTime',
    },
    {
      title: '收藏状态',
      dataIndex: 'is_fav',
      valueType: 'text',
      hideInTable: true,
      valueEnum: arrayToProTableColumns(isFav),
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <div
          style={{
            width: '100px',
          }}
        >
          <a
            onClick={() => {
              openRunningInfoLog(record);
            }}
          >
            日志
          </a>
          {/* <Divider type="vertical" /> */}
          {/* <DeleteConfirm record={record}  removeRule={removeWork}  /> */}
          {/* <a
           onClick={() => {
            deleteConfirm(record);
          }}

          >删除</a> */}
        </div>
      ),
    },
  ];

  return (
    <PageHeaderWrapper title={false}>
      <ProTable
        // headerTitle="查询表格"
        actionRef={actionRef}
        onSearch={(keyword) => {
          console.log('搜索', keyword);
        }}
        request={(params) => getActivities({ ...params })}
        columns={columns}
        rowKey="id"
        search={{
          collapsed: false,
          optionRender: ({ searchText }, { form }) => {
            return (
              <>
                <Button
                  type="primary"
                  onClick={() => {
                    form.submit();
                  }}
                >
                  {searchText}
                </Button>
              </>
            );
          },
        }}
        // search={false}
      />
    </PageHeaderWrapper>
  );
};

export default TableList;
