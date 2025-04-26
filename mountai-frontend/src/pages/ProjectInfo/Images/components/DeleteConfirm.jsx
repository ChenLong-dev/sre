import { Button, Divider, Dropdown, Menu, message, Input, Popconfirm } from 'antd';
import React from 'react';

const DeleteConfirm = ({record, removeRule}) => {
  function confirm(e) {
    console.log(e);
    removeRule(record)
  //  message.success('删除成功');
  }
  
  function cancel(e) {
    message.error('取消删除');
  }  

  return (
    <Popconfirm
      title="确认删除吗?"
      onConfirm={confirm}
      onCancel={cancel}
      okText="Yes"
      cancelText="No"
    >
      <a href="#">删除</a>
    </Popconfirm>
  );
}


export default DeleteConfirm;