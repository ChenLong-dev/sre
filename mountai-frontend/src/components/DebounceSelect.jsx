import React, { useEffect } from 'react';
import { message, Select, Spin } from 'antd';
import { debounce } from 'lodash';

function DebounceSelect({ getOptions, debounceTimeout = 800, ...props }) {
  const [fetching, setFetching] = React.useState(false);
  const [options, setOptions] = React.useState([]);
  const fetchRef = React.useRef(0);
  const debounceFetcher = React.useMemo(() => {
    const loadOptions = async (value) => {
      fetchRef.current += 1;
      const fetchId = fetchRef.current;
      setOptions([]);
      setFetching(true);
      try {
        const newOptions = await getOptions(value);
        if (fetchId !== fetchRef.current) {
          return;
        }

        setOptions(newOptions);
        setFetching(false);
      } catch (error) {
        message.error(error.message);
      }
    };

    return debounce(loadOptions, debounceTimeout);
  }, [getOptions, debounceTimeout]);

  useEffect(() => {
    debounceFetcher('');
    return () => {
      fetchRef.current = 0;
    };
  }, []);

  return (
    <Select
      filterOption={false}
      onSearch={debounceFetcher}
      notFoundContent={fetching ? <Spin size="small" /> : null}
      {...props}
      options={options}
    />
  );
}

export default React.memo(DebounceSelect);
