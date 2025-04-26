import React, { useCallback } from 'react';
import { message } from 'antd';
import { map } from 'lodash';
import getUsers from '@/services/getUsers';
import DebounceSelect from './DebounceSelect';

function SelectUsers({ ...props }) {
  const getUserList = useCallback(
    async (keyword) => {
      try {
        const { list: userList } = await getUsers(keyword);
        return map(userList, (user) => ({
          label: user.name,
          value: user.id,
        }));
      } catch (error) {
        message.error(error.message);
      }
      return '';
    },
    [getUsers]
  );

  return (
    <DebounceSelect
      {...props}
      placeholder="请选择负责人"
      getOptions={getUserList}
    />
  );
}

export default React.memo(SelectUsers);
