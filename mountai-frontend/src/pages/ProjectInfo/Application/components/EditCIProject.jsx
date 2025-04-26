import React, { useCallback, useEffect, useState, useMemo } from 'react';
import PropTypes from 'prop-types';
import {
  message,
  Typography,
  Modal,
  Form,
  Checkbox,
  Switch,
  Select,
  Button,
} from 'antd';
import { queryGitBranch } from '@/services/git';
import { map, concat } from 'lodash';
import {
  PIPELINE_STAGE_ARRAY,
  PIPELINE_STAGE_OPTIONS,
} from '../constants/pipelineStages';
import {
  getProjectCIStatus,
  setProjectCI,
  putProjectCIStatus,
} from '../service';

const { Item } = Form;

const defaultValues = {
  message_notification: ['email', 'dingding'],
  allow_merge_switch: false,
  pipeline_stages: PIPELINE_STAGE_ARRAY,
  deploy_branch_name: { fat: '', test: '' },
};

const EditCIProject = ({ data, onOk: oo, ...props }) => {
  const [form] = Form.useForm();
  const [isEdit, setIsEdit] = useState(false);
  const [projectCIStatus, setProjectCIStatus] = useState({});
  const [branches, setBranches] = useState([]);
  const [showNodeLabels, setShowNodeLabels] = useState(false);
  const initialValues = useMemo(() => projectCIStatus || defaultValues, [
    projectCIStatus,
  ]);

  const onOk = useCallback(async () => {
    const value = await form.validateFields();
    try {
      if (isEdit) {
        await putProjectCIStatus(data.id, value);
      } else {
        await setProjectCI(data.id, value);
      }
      setIsEdit(true);
      message.success(`${isEdit ? '修改' : '创建'}CI流程成功`);
      oo();
    } catch (e) {
      message.error(e.message || 'CI流程失败');
    }
  }, [form, data, isEdit]);

  useEffect(() => {
    (async () => {
      if (data) {
        try {
          const [
            projectCIStatusResult,
            branchesResult,
          ] = await Promise.allSettled([
            getProjectCIStatus(data.id),
            queryGitBranch(data.id, null),
          ]);
          if (projectCIStatusResult.value) {
            setProjectCIStatus(projectCIStatusResult.value);
            setIsEdit(true);
          } else {
            setProjectCIStatus({});
          }
          if (branchesResult.value) {
            setBranches(
              concat(
                [{ label: '当前分支', value: '' }],
                map(branchesResult.value, (branch) => ({
                  label: branch.name,
                  value: branch.name,
                }))
              )
            );
          }
        } catch (e) {
          message.error(e.message);
        }
      }
    })();
  }, [data]);
  return (
    <Modal
      {...props}
      title={isEdit ? `项目${data.name} ci流程已创建（编辑）` : '创建'}
      okText="确认"
      cancelText="取消"
      onOk={onOk}
      destroyOnClose
    >
      {isEdit && (
        <div>
          url:&nbsp;
          <Typography.Paragraph
            copyable={{
              text: `${projectCIStatus?.view_url}`,
            }}
            style={{ display: 'inline' }}
          >
            <a
              href={`${projectCIStatus?.view_url}`}
              rel="noopener noreferrer"
              target="__blank"
            >
              {projectCIStatus?.view_url}
            </a>
          </Typography.Paragraph>
        </div>
      )}
      <Form form={form} initialValues={initialValues}>
        <Item name="message_notification" label="ci消息通知方式">
          <Checkbox.Group>
            <Checkbox value="email">邮箱</Checkbox>
            <Checkbox value="dingding">钉钉</Checkbox>
          </Checkbox.Group>
        </Item>
        <Item
          name="allow_merge_switch"
          label="ci不通过不允许合并代码"
          valuePropName="checked"
        >
          <Switch />
        </Item>
        <Item label="部署分支配置">
          <Item name={['deploy_branch_name', 'fat']} label="fat">
            <Select
              placeholder="请选择 fat"
              showSearch
              options={branches}
              optionFilterProp="label"
            />
          </Item>
          <Item name={['deploy_branch_name', 'test']} label="test">
            <Select
              placeholder="请选择 test"
              showSearch
              options={branches}
              optionFilterProp="label"
            />
          </Item>
        </Item>
        {/* 高级选项 */}
        <Item label="高级选项">
          <Button
            onClick={() => setShowNodeLabels((prevState) => !prevState)}
            type="primary"
            size="small"
          >
            {showNodeLabels ? '收起' : '修改'}
          </Button>
        </Item>
        <Item
          name="pipeline_stages"
          label="工作流阶段"
          style={showNodeLabels ? { display: 'block' } : { display: 'none' }}
        >
          <Select
            mode="multiple"
            showArrow
            placeholder="请选择工作流阶段"
            options={PIPELINE_STAGE_OPTIONS}
          />
        </Item>
      </Form>
    </Modal>
  );
};

EditCIProject.propTypes = {
  data: PropTypes.shape({
    id: PropTypes.string,
    name: PropTypes.string,
  }).isRequired,
  onOk: PropTypes.func.isRequired,
  onCancel: PropTypes.func.isRequired,
};

export default React.memo(EditCIProject);
