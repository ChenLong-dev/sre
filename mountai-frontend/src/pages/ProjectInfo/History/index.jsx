import { Button, Divider, message, Modal } from 'antd';
import React, { useState, useRef } from 'react';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import WordTip from '@/components/Common/WordTip';

// 应用类型  应用服务类型
import { applicationType, deployActions } from '@/utils/commonData';

import { isEmpty, isEqual, pick, reduce } from 'lodash';
import Styles from './index.less';
import { queryActivities } from './service';

const arrayToProTableColumns = (datas, keyname) => {
  return reduce(
    datas,
    (result, { value, name }) => {
      return {
        ...result,
        [value]: { text: name, [keyname]: name },
      };
    },
    {}
  );
};

const TableList = (props) => {
  const {
    location: { query },
    match: {
      params: { projectId },
    },
  } = props;

  const [newLoad, setNewLoad] = useState(true);
  const actionRef = useRef();

  const getActivities = ({ ...params }) => {
    if (newLoad) {
      setNewLoad(false);
      return queryActivities({ ...params, ...query, projectId });
    }

    const newquery = pick(params, [
      'env_name',
      'action',
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
    return queryActivities({ ...newquery, projectId });
  };

  /**
   *  打开运行日志
   */
  const openRunningInfoLog = (record) => {
    console.log('日志打印:', record);

    if (record.detail && typeof record.detail === 'string') {
      Modal.info({
        title: '日志',
        width: 950,
        maskClosable: true,
        content: <div className={Styles.preLine}>{record.detail}</div>,
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
      hideInForm: true,
    },
    {
      title: '操作类型',
      dataIndex: 'action',
      initialValue: query?.action || '',
      valueEnum: arrayToProTableColumns(deployActions, 'action'),
      hideInTable: true,
    },
    // {
    //   title: 'ID',
    //   dataIndex: 'id',
    //   hideInTable:true,
    // },
    {
      title: '版本',
      dataIndex: 'version',
      width: 300,
      hideInSearch: true,
      render: (text, record) => (
        <div className={Styles.version}>
          <div>{text}</div>
          {record.image_version && (
            <WordTip title={record.image_version} className={Styles.icon} />
          )}
        </div>
      ),
    },
    {
      title: '应用类型',
      dataIndex: 'app_type',
      initialValue: query && query.app_type ? query.app_type : '',
      valueType: 'text',
      // dataIndex: 'teamid',
      valueEnum: arrayToProTableColumns(applicationType, 'app_type'),
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
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <>
          <a
            onClick={() => {
              openRunningInfoLog(record);
            }}
          >
            日志
          </a>
          <Divider type="vertical" />
          {/* <DeleteConfirm record={record}  removeRule={removeWork}  /> */}
          {/* <a
           onClick={() => {
            deleteConfirm(record);
          }}

          >删除</a> */}
        </>
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
      />
    </PageHeaderWrapper>
  );
};

export default TableList;
