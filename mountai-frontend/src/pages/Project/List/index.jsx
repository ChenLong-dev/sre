import { PlusOutlined, StarFilled } from '@ant-design/icons';
import { Button, Divider, message, Tag } from 'antd';
import React, { useState, useRef, useEffect } from 'react';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import DeleteConfirm from '@/components/DeleteConfirm';
import { queryAllTeams } from '@/services/team';
import { isEmpty, isEqual, join, pick, reduce, split, map } from 'lodash';
import SelectUsers from '@/components/SelectUsers';
import {
  queryProject,
  removeProject,
  getAllProjectLabels,
  collectProject,
  uncollectProject,
} from './service';

const TableList = (props) => {
  let tableForm = null;
  const {
    location: { query },
  } = props;
  const [teams, setTeams] = useState({});
  const [labels, setLabels] = useState({});
  // const [search, setSearch] = useState({});
  const [newLoad, setNewLoad] = useState(true);

  // 收藏/取消收藏项目
  const handleColllectProject = async (record) => {
    if (!record.is_fav) {
      try {
        await collectProject(record.id);
        message.success('已收藏');
        tableForm.submit();
      } catch (error) {
        message.error(error.message || '收藏失败');
      }
    } else {
      try {
        await uncollectProject(record.id);
        message.success('已取消收藏');
        tableForm.submit();
      } catch (error) {
        message.error(error.message || '取消收藏失败');
      }
    }
  };

  const actionRef = useRef();

  const removeWork = async (record) => {
    try {
      await removeProject(record);
      message.success('删除成功');
      if (actionRef.current) {
        actionRef.current.reload();
      }
    } catch (error) {
      message.error(error.message || '删除失败');
    }
  };

  const columns = [
    {
      title: '项目名称',
      dataIndex: 'name',
      valueType: 'text',
      hideInSearch: true,
      rules: [
        {
          required: true,
          message: '规则名称为必填项',
        },
      ],
      render: (_, record) => (
        <>
          <a href={`/project/${record.id}/application`}>{record.name}</a>
        </>
      ),
    },
    {
      title: '关键词搜索',
      dataIndex: 'keyword',
      initialValue: query && query.keyword ? query.keyword : '',
      hideInTable: true,
      hideInForm: true,
      rules: [
        {
          message: '请输入关键词',
        },
      ],
    },
    {
      title: '项目语言',
      valueType: 'text',
      dataIndex: 'language',
      initialValue: query && query.language ? query.language : '',
      valueEnum: {
        Go: {
          text: 'GO',
          language: 'Go',
        },
        JavaScript: {
          text: 'NodeJS',
          language: 'JavaScript',
        },
        Lua: {
          text: 'Lua',
          language: 'Lua',
        },
        Python: {
          text: 'Python',
          language: 'Python',
        },
        PHP: {
          text: 'PHP',
          language: 'PHP',
        },
        Others: {
          text: 'Other',
          language: 'Others',
        },
      },
    },
    {
      title: '项目描述',
      valueType: 'textarea',
      dataIndex: 'desc',
      // sorter: true,
      hideInForm: true,
      hideInSearch: true,
      // renderText: val => `${val}`,
      render: (text, record) => (
        <div
          style={{
            width: '150px',
            overflow: 'hidden',
            textOverflow: 'ellipsis',
            whiteSpace: 'nowrap',
            wordWrap: 'break-word',
            wordBreak: 'normal',
          }}
        >
          {record.desc}
        </div>
      ),
    },
    {
      title: '项目标签',
      dataIndex: 'labels',
      hideInForm: true,
      initialValue: query?.labels ? split(query.labels, ',') : undefined,
      valueEnum: labels,
      fieldProps: {
        mode: 'multiple',
      },
      render: (_, record) => (
        <div
          style={{
            width: '100px',
          }}
        >
          {record.labels.map((item, index) => (
            <Tag color="blue" key={index}>
              {labels[item]?.text || null}
            </Tag>
          ))}
        </div>
      ),
    },
    {
      title: '负责人',
      valueType: 'text',
      dataIndex: 'owner_id',
      hideInForm: true,
      initialValue: query?.owner_id ?? '',
      render: (_, record) => (
        <>{join(map(record.owners, ({ name }) => name))}</>
      ),
      renderFormItem: (_, { type, ...rest }) => {
        if (type === 'form') {
          return null;
        }
        return <SelectUsers {...rest} showSearch />;
      },
    },
    {
      title: '项目团队',
      valueType: 'text',
      dataIndex: 'team_id',
      // sorter: true,
      hideInForm: true,
      initialValue: query?.team_id ?? '',
      valueEnum: teams,
      render: (_, record) => <>{record.team.name}</>,
    },
    {
      title: '更新时间',
      hideInSearch: true,
      dataIndex: 'update_time',
      // sorter: true,
      valueType: 'dateTime',
      hideInForm: true,
    },
    {
      title: '收藏状态',
      render: (_, record) => (
        <StarFilled
          onClick={() => handleColllectProject(record)}
          style={{
            fontSize: '18px',
            cursor: 'pointer',
            color: record.is_fav ? '#fadb14' : '',
          }}
        />
      ),
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <div
          style={{
            width: '150px',
          }}
        >
          <a href={`/project/${record.id}/application`}>详情</a>
          <Divider type="vertical" />
          <a href={`/projects/update/${record.id}`}>编辑</a>
          <Divider type="vertical" />

          <DeleteConfirm record={record} removeRule={removeWork} />
        </div>
      ),
    },
  ];

  useEffect(() => {
    queryAllTeams().then((res) => {
      setTeams(
        reduce(
          res.list,
          (result, { name, id }) => ({
            ...result,
            [id]: { text: name, team_id: id },
          }),
          {}
        )
      );
    });
    getAllProjectLabels().then((res) => {
      setLabels(
        reduce(
          res,
          (result, { name, label }) => ({
            ...result,
            [label]: { text: name },
          }),
          {}
        )
      );
    });
  }, []);

  const getProject = (params) => {
    if (newLoad) {
      setNewLoad(false);
      return queryProject({ ...params, ...query });
    }
    const newquery = pick(
      {
        ...params,
        labels: join(params?.labels),
      },
      [
        'team_id',
        'owner_id',
        'language',
        'keyword',
        'labels',
        'pageSize',
        'current',
      ]
    );
    if (!isEqual(newquery, query)) {
      if (!isEmpty(newquery)) {
        const Url = new URLSearchParams(newquery);
        window.history.replaceState(null, null, `?${Url.toString()}`);
      } else {
        window.history.replaceState(null, null, `?`);
      }
    }
    return queryProject({ ...newquery });
  };

  return (
    <PageHeaderWrapper>
      <ProTable
        actionRef={actionRef}
        rowKey="id"
        search={{
          collapsed: false,
          optionRender: ({ searchText }, { form }) => {
            tableForm = form;
            document.onkeydown = (e) => {
              const evt = e;
              if (evt.keyCode === 13) {
                form.submit();
              }
            };

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
        toolBarRender={() => [
          <a href="/projects/create">
            <Button type="primary">
              <PlusOutlined /> 新建
            </Button>
          </a>,
        ]}
        onReset={() => {
          actionRef.current.clearSelected();
        }}
        request={(params) => getProject(params)}
        columns={columns}
      />
    </PageHeaderWrapper>
  );
};

export default TableList;
