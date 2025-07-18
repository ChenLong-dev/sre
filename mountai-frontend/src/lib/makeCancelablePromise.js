var _interopRequireDefault = require("@babel/runtime/helpers/interopRequireDefault");

var _classCallCheck2 = _interopRequireDefault(require("@babel/runtime/helpers/classCallCheck"));

var _assertThisInitialized2 = _interopRequireDefault(require("@babel/runtime/helpers/assertThisInitialized"));

var _inherits2 = _interopRequireDefault(require("@babel/runtime/helpers/inherits"));

var _possibleConstructorReturn2 = _interopRequireDefault(require("@babel/runtime/helpers/possibleConstructorReturn"));

var _getPrototypeOf2 = _interopRequireDefault(require("@babel/runtime/helpers/getPrototypeOf"));

var _wrapNativeSuper2 = _interopRequireDefault(require("@babel/runtime/helpers/wrapNativeSuper"));

var _defineProperty2 = _interopRequireDefault(require("@babel/runtime/helpers/defineProperty"));

function _createSuper(Derived) { var hasNativeReflectConstruct = _isNativeReflectConstruct(); return function _createSuperInternal() { var Super = (0, _getPrototypeOf2["default"])(Derived), result; if (hasNativeReflectConstruct) { var NewTarget = (0, _getPrototypeOf2["default"])(this).constructor; result = Reflect.construct(Super, arguments, NewTarget); } else { result = Super.apply(this, arguments); } return (0, _possibleConstructorReturn2["default"])(this, result); }; }

function _isNativeReflectConstruct() { if (typeof Reflect === "undefined" || !Reflect.construct) return false; if (Reflect.construct.sham) return false; if (typeof Proxy === "function") return true; try { Boolean.prototype.valueOf.call(Reflect.construct(Boolean, [], function () {})); return true; } catch (e) { return false; } }

var CanceledError = function (_Error) {
  (0, _inherits2["default"])(CanceledError, _Error);

  var _super = _createSuper(CanceledError);

  function CanceledError() {
    var _this;

    (0, _classCallCheck2["default"])(this, CanceledError);
    _this = _super.call(this, 'promise is been canceled');
    (0, _defineProperty2["default"])((0, _assertThisInitialized2["default"])(_this), "canceled", true);
    return _this;
  }

  return CanceledError;
}((0, _wrapNativeSuper2["default"])(Error));

export function makeCancelablePromise(promise) {
  var hasCanceled = false;
  var wrappedPromise = new Promise(function (resolve, reject) {
    promise.then(function (val) {
      return hasCanceled ? reject(new CanceledError()) : resolve(val);
    }, function (error) {
      return hasCanceled ? reject(new CanceledError()) : reject(error);
    });
  });

  wrappedPromise.cancel = function () {
    hasCanceled = true;
  };

  return wrappedPromise;
}