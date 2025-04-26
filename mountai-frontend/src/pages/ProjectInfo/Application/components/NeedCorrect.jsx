import React, { useState } from 'react';
import { Form, Button, DatePicker, Input, Modal, Radio, Select, Steps } from 'antd';

const FormItem = Form.Item;
const { Step } = Steps;
const { TextArea } = Input;
const { Option } = Select;
const RadioGroup = Radio.Group;
const formLayout = {
  labelCol: {
    span: 7,
  },
  wrapperCol: {
    span: 13,
  },
};

const NeedCorrect = props => {

  const [form] = Form.useForm();

  const {
    onSubmit: handleUpdate,
    updateApplication,
    onCancel: handleUpdateModalVisible,
    modalVisible,
    values,
    application,
  } = props;

  const submitBuild = async () => {
    const fieldsValue = await form.validateFields();
    console.log('fieldsValue', fieldsValue);
    fieldsValue.appid = application.id;
    handleUpdate(fieldsValue); // 更新应用
  };

 
  const renderFooter = () => {
    return (
      <>
       <Button onClick={() => handleUpdateModalVisible(false, values)}>取消</Button>
        <Button type="primary" onClick={() => submitBuild()}>
          提交
        </Button>
      </>
    );
  };  


  return (
    <Modal
      width={640}
      bodyStyle={{
        padding: '32px 40px 48px',
      }}
      destroyOnClose
      title="应用更新"
      visible={modalVisible}
      footer={renderFooter()}
      onCancel={() => handleUpdateModalVisible()}
    >
    
      <Form
        {...formLayout}
        form={form}
        initialValues={{
        }}
      >

        <div style={{color:'red',  textAlign:'center',  margin: '20px auto' }}> 
          {/* 旧系统应用首次发布需迁移并重启，迁移请先录入应用名称  */}
          请注意：修改应用名后，包括线上环境在内的所有环境都会以上次部署参数自动重新部署！
        </div>
       
        <FormItem
          name="name"
          label="新名称"
          // rules={[{ required: true, message: '不能为空' }]}
          rules={[
            {
              required: true,
              message: '应用名为必填项',
            },
            ({ getFieldValue }) => ({
              validator(rule, value) {
                let pattern = new RegExp("[\u4E00-\u9FA5]+");
                if(pattern.test(value)) {
                  return Promise.reject('应用名不能为汉字');
                }
                let reg = new RegExp(/^[A-Za-z0-9\-]+$/); 
              
                if (value ) {
                  if(!reg.test(value)) {
                    return Promise.reject('应用名不合法');
                  }
                  return Promise.resolve();
                }
                
              },
            }),
          ]}
        >
          
          <Input 
          placeholder="名称应简短,如:restful/order-worker/dump-job等" 
          />
        </FormItem>
      </Form>
    </Modal>
  );
};

export default NeedCorrect;
