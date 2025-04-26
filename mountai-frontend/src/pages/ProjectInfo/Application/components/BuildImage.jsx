import React, { useState, useEffect } from 'react';
import { QuestionCircleOutlined } from '@ant-design/icons';
import {
  Form,
  Tooltip,
  Input,
  Modal,
  Select,
  message,
  Spin,
  Radio,
} from 'antd';
import { queryGitBranch } from '@/services/git';
import { debounce, forEach } from 'lodash';

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

const UpdateForm = ({
  onSubmit: handleUpdate,
  onCancel: handleUpdateModalVisible,
  modalVisible,
  projectId,
  imageArgs,
  imageArgTemplates,
}) => {
  const [formVals, setFormVals] = useState({
    projectId,
    branch: [],
  });
  const [form] = Form.useForm();
  const [commitMsg, setCommitMsg] = useState();
  const [branchLoading, setBranchLoading] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);

  useEffect(() => {
    if (modalVisible) {
      form.resetFields();
      setBranchLoading(true);
      queryGitBranch(projectId, null)
        .then((res) => {
          setFormVals({
            projectId,
            branch: res,
          });
          setBranchLoading(false);
        })
        .catch((err) => {
          message.error(new Error(err));
        });
    }
  }, [modalVisible]);

  const submitBuild = async () => {
    try {
      setConfirmLoading(true);
      const fieldsValue = await form.validateFields();
      setFormVals({ ...formVals, ...fieldsValue });
      await handleUpdate({ ...formVals, ...fieldsValue });
    } finally {
      setConfirmLoading(false);
    }
  };

  const onSearch = (val) => {
    setBranchLoading(true);
    // 调用
    queryGitBranch(projectId, val)
      .then((res) => {
        setFormVals({
          projectId,
          branch: res,
        });
        setBranchLoading(false);
      })
      .catch((err) => {
        message.error(new Error(err));
      });
  };

  const branchChange = (value) => {
    if (!formVals.projectId || !value) {
      return;
    }

    // 新构建参数
    form.setFieldsValue({
      build_arg: imageArgs && imageArgs[value] ? imageArgs[value] : '',
    });

    // getProjectSuccessImagesInfo(formVals.projectId, value).then( res=> {
    //     if(res && res.length > 0) {
    //       form.setFieldsValue({
    //         build_arg: res[0].build_arg ? res[0].build_arg:'',
    //       });
    //     }
    // });

    forEach(formVals.branch, (item) => {
      if (item.name === value) {
        if (item.commit) {
          form.setFieldsValue({
            commit_id: item.commit.short_id,
          });
          setCommitMsg(item?.commit?.message || '');
        }
      }
    });
  };

  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title="构建镜像"
      visible={modalVisible}
      onCancel={() => {
        setFormVals({
          projectId,
          branch: [],
        });
        handleUpdateModalVisible();
      }}
      onOk={submitBuild}
      okText="创建"
      confirmLoading={confirmLoading}
    >
      <Form
        {...formLayout}
        form={form}
        initialValues={{
          projectId,
          // ding_hook: formVals.ding_hook,
          // type: formVals.type,
          build_arg_type: 'custom',
          build_arg: '',
        }}
      >
        <FormItem
          // name="projectId"
          label="Git Id"
        >
          {projectId}
        </FormItem>
        <FormItem
          name="branch_name"
          label="分支"
          rules={[{ required: true, message: '分支不能为空' }]}
        >
          <Select
            showSearch
            onChange={branchChange}
            onSearch={debounce(onSearch, 800)}
            notFoundContent={branchLoading ? <Spin size="small" /> : null}
          >
            {formVals.branch &&
              formVals.branch.map((item, index) => {
                return (
                  <Option key={index} value={item.name}>
                    {' '}
                    {item.name}{' '}
                  </Option>
                );
              })}
          </Select>
        </FormItem>

        <FormItem
          name="commit_id"
          label={
            <>
              commit_id&nbsp;
              <Tooltip title={commitMsg || ''} placement="topRight">
                <QuestionCircleOutlined />
              </Tooltip>
            </>
          }
        >
          <Input placeholder="" disabled />
        </FormItem>
        <FormItem name="build_arg_type" label="构建参数类型">
          <Radio.Group>
            <Radio value="custom">自定义</Radio>
            <Radio value="template">模板</Radio>
          </Radio.Group>
        </FormItem>
        <FormItem
          noStyle
          shouldUpdate={(prevValues, curValues) =>
            prevValues.build_arg_type !== curValues.build_arg_type
          }
        >
          {({ getFieldValue }) => {
            const type = getFieldValue('build_arg_type');
            if (type === 'custom') {
              return (
                <FormItem name="build_arg" label="自定义构建参数">
                  <Input />
                </FormItem>
              );
            }
            if (type === 'template') {
              return (
                <FormItem name="build_args_template_id" label="构建参数模板">
                  <Select>
                    {imageArgTemplates &&
                      imageArgTemplates.map((item) => {
                        return (
                          <Option key={item.id} value={item.id}>
                            {item.name}
                          </Option>
                        );
                      })}
                  </Select>
                </FormItem>
              );
            }

            return null;
          }}
        </FormItem>
        <FormItem name="description" label="描述">
          <Input.TextArea />
        </FormItem>
      </Form>
    </Modal>
  );
};

export default UpdateForm;
