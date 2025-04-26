import React from 'react';

export function useEventCallback(fn) {
  var ref = (0, React.useRef)(fn);
  (0, React.useLayoutEffect)(function () {
    ref.current = fn;
  });
  return (0, React.useCallback)(function () {
    for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
      args[_key] = arguments[_key];
    }

    return ref.current.apply(0, args);
  }, []);
}