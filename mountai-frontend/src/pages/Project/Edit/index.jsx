import React, { useState, useEffect } from 'react';
import { Button, Card, Input, Form, Select, message, Checkbox } from 'antd';
import { FormattedMessage, history, useLocation, useParams } from 'umi';
import { PageHeaderWrapper } from '@ant-design/pro-layout';
import { languageConfig } from '@/utils/commonData';
import { map } from 'lodash';
import SelectUsers from '@/components/SelectUsers';
import styles from './style.less';
import {
  getGitInfobyid,
  getTeams,
  getProjectInfo,
  updateProject,
  postProject,
  getAllProjectLabels,
} from './service';

const FormItem = Form.Item;
const { Option } = Select;
const { TextArea } = Input;

const Edit = () => {
  const { projectId } = useParams();
  const { query } = useLocation();
  const [form] = Form.useForm();
  const [teams, setTeams] = React.useState([]);
  const [allLabels, setAllLabels] = useState([]);
  const [gitValidatorInfo, setGitValidatorInfo] = useState({
    validateStatus: null,
    help: null,
  });

  const formItemLayout = {
    labelCol: {
      xs: {
        span: 24,
      },
      sm: {
        span: 7,
      },
    },
    wrapperCol: {
      xs: {
        span: 24,
      },
      sm: {
        span: 12,
      },
      md: {
        span: 10,
      },
    },
  };
  const submitFormLayout = {
    wrapperCol: {
      xs: {
        span: 24,
        offset: 0,
      },
      sm: {
        span: 10,
        offset: 7,
      },
    },
  };

  const onFinish = async (values) => {
    if (projectId) {
      try {
        await updateProject({
          ...values,
          id: projectId,
        });
        message.success('项目保存成功');

        if (query && query.redirect) {
          history.push(query.redirect);
        } else {
          history.push('/projects/list');
        }
      } catch (error) {
        message.error(error.message || '项目保存失败');
      }
    } else {
      try {
        const result = await postProject(values);
        message.success('项目创建成功');
        history.push(`/project/${result?.id}/application`);
      } catch (error) {
        message.error(error.message || '项目创建失败');
      }
    }
  };

  const onFinishFailed = (errorInfo) => {
    // eslint-disable-next-line no-console
    console.log('Failed:', errorInfo);
  };

  const onValuesChange = async (changedValues) => {
    const { git_id: gitId } = changedValues;
    if (gitId) {
      setGitValidatorInfo({ validateStatus: null, help: null });
    }
  };

  useEffect(() => {
    getTeams()
      .then((res) => {
        setTeams(res.list || []);
      })
      .catch(() => setTeams([]));
    getAllProjectLabels().then((res) => setAllLabels(res || []));
    if (projectId) {
      getProjectInfo(projectId).then((res) => {
        const {
          name,
          language,
          desc,
          team,
          api_doc_url: apiDocUrl,
          dev_doc_url: devDocUrl,
          labels,
          owners,
        } = res;
        form.setFieldsValue({
          name,
          desc,
          language,
          team_id: team.id,
          api_doc_url: apiDocUrl,
          dev_doc_url: devDocUrl,
          labels,
          owner_ids: map(owners, ({ id }) => id),
        });
      });
    }
  }, []);

  const onGit = async () => {
    const gitId = form.getFieldValue('git_id');

    if (gitId && gitId.length >= 1) {
      const data = await getGitInfobyid(gitId);
      if (!data || !data.web_url) {
        setGitValidatorInfo({
          validateStatus: 'error',
          help: '没有获取到git信息',
        });
      } else {
        form.setFieldsValue({
          name: data.path,
        });
      }
    }
  };

  return (
    <PageHeaderWrapper title={false}>
      <Card bordered={false}>
        <Form
          hideRequiredMark
          style={{
            marginTop: 8,
          }}
          form={form}
          name="basic"
          onFinish={onFinish}
          onFinishFailed={onFinishFailed}
          onValuesChange={onValuesChange}
          {...formItemLayout}
        >
          {projectId ? (
            <FormItem label="Git ID">{projectId}</FormItem>
          ) : (
            <FormItem
              label="Git ID"
              name="git_id"
              onBlur={onGit}
              validateStatus={gitValidatorInfo.validateStatus}
              help={gitValidatorInfo.help}
              rules={[
                {
                  required: true,
                  message: '项目对应的Git ID为必填项',
                },
              ]}
            >
              <Input placeholder="请输入项目对应的Git ID" />
            </FormItem>
          )}

          <FormItem
            label="项目名称"
            name="name"
            rules={[
              {
                required: true,
                message: '项目名称为必填项',
              },
              () => ({
                validator(_, value) {
                  const pattern = new RegExp('[\u4E00-\u9FA5]+');

                  if (pattern.test(value)) {
                    return Promise.reject(new Error('项目名称不能为汉字'));
                  }
                  const reg = new RegExp(/^[A-Za-z0-9\\-]+$/);

                  if (value && !reg.test(value)) {
                    return Promise.reject(new Error('项目名称不合法'));
                  }
                  return Promise.resolve();
                },
              }),
            ]}
          >
            <Input placeholder="请输入项目名称" disabled={!!projectId} />
          </FormItem>

          <FormItem
            label="项目语言"
            name="language"
            rules={[
              {
                required: true,
                message: '项目语言为必填项',
              },
            ]}
          >
            <Select
              placeholder="请选择项目语言"
              disabled={!!projectId}
              allowClear
            >
              {languageConfig.map((item, index) => {
                return (
                  <Option key={index} value={item.value}>
                    {item.text}
                  </Option>
                );
              })}
            </Select>
          </FormItem>

          <FormItem
            label="所属团队"
            name="team_id"
            rules={[
              {
                required: true,
                message: '请选择团队',
              },
            ]}
          >
            <Select placeholder="请选择团队" allowClear>
              {teams.map((item, index) => {
                return (
                  <Option key={index} value={item.id}>
                    {item.name}
                  </Option>
                );
              })}
            </Select>
          </FormItem>

          <FormItem label="项目标签" name="labels">
            <Select mode="multiple" placeholder="请选择标签" allowClear>
              {allLabels.map((item, index) => {
                return (
                  <Option key={index} value={item.label}>
                    {item.name}
                  </Option>
                );
              })}
            </Select>
          </FormItem>

          <FormItem
            label="负责人"
            name="owner_ids"
            rules={[
              {
                required: true,
                message: '请选择负责人',
              },
            ]}
          >
            <SelectUsers mode="multiple" />
          </FormItem>

          <FormItem
            label={
              <span>
                项目描述
                <em className={styles.optional}>
                  <FormattedMessage id="projectandcreate.form.optional" />
                </em>
              </span>
            }
            name="desc"
          >
            <TextArea
              style={{
                minHeight: 32,
              }}
              placeholder="请输入项目描述"
              rows={4}
            />
          </FormItem>
          <FormItem
            label={
              <span>
                开发文档
                <em className={styles.optional}>
                  <FormattedMessage id="projectandcreate.form.optional" />
                </em>
              </span>
            }
            name="dev_doc_url"
          >
            <Input placeholder="请输入项目的开发文档地址" />
          </FormItem>

          <FormItem
            label={
              <span>
                接口文档
                <em className={styles.optional}>
                  <FormattedMessage id="projectandcreate.form.optional" />
                </em>
              </span>
            }
            name="api_doc_url"
          >
            <Input placeholder="请输入项目的API文档地址" />
          </FormItem>

          {!projectId && (
            <FormItem
              label={
                <span>
                  是否禁用CI
                  <em className={styles.optional}>
                    <FormattedMessage id="projectandcreate.form.optional" />
                  </em>
                </span>
              }
              name="disable_ci"
              valuePropName="checked"
            >
              <Checkbox />
            </FormItem>
          )}
          <FormItem
            {...submitFormLayout}
            style={{
              marginTop: 32,
            }}
          >
            <Button type="primary" htmlType="submit">
              {projectId ? '确认修改' : '确认创建'}
            </Button>
          </FormItem>
        </Form>
      </Card>
    </PageHeaderWrapper>
  );
};

export default React.memo(Edit);
