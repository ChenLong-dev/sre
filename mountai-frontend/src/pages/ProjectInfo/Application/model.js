import { message } from 'antd';
import {
  getProjectInfo,
  getProjectImagesInfo,
  getProjectApplication,
  getProjectimageTag,
  getAllProjectLabels,
  getMemberRole,
  getProjectClusters,
} from './service';
import { getTemplates } from '../../Team/ImageArgTemplate/service';

const initialValues = {
  images: [],
  projectInfo: null,
  applications: [],
  imageTags: [],
  allApplications: [],
  imageTagsOption: [],
  allLabels: [],
  imageArgTemplates: [],
  projectClusterMap: {},
};

const Model = {
  namespace: 'application',
  state: initialValues,
  effects: {
    *init({ payload }, { put }) {
      //   console.log('_.payload-->', _.payload);

      yield put({
        type: 'getProjectInfo',
        payload: payload.project_id,
      });

      // 性能考虑 解除强制刷新 交给组件自行加载service层
      // yield put({
      //   type: 'getProjectImagesInfo',
      //   payload: _.payload
      // });

      yield put({
        type: 'getProjectApplication',
        payload,
      });

      // yield put({
      //   type: 'getProjectimageTag',
      //   payload: _.payload
      // });

      yield put({
        type: 'getAllLabels',
        payload,
      });

      yield put({
        type: 'getMemberRole',
        payload: payload.project_id,
      });
    },
    *initAll({ payload }, { call, put }) {
      const response = yield call(getProjectApplication, payload);
      yield put({
        type: 'saveAll',
        payload: { allApplications: response.list || [] },
      });
    },
    *getProjectInfo(_, { call, put }) {
      const response = yield call(getProjectInfo, _.payload);
      // console.log('model.application.*getProjectInfo()=>', _.payload);
      yield put({
        type: 'save',
        payload: {
          projectInfo: response,
        },
      });
      yield put({
        type: 'getImageArgTemplate',
        payload: {
          teamId: response.team.id,
        },
      });
    },
    *getProjectImagesInfo(_, { call, put }) {
      const { list } = yield call(getProjectImagesInfo, _.payload);
      // console.log('model.application.*getProjectImagesInfo()=>',response);
      yield put({
        type: 'save',
        payload: {
          images: list,
        },
      });
    },

    *getProjectApplication(_, { call, put }) {
      const response = yield call(getProjectApplication, _.payload);

      yield put({
        type: 'save',
        payload: {
          applications: response.list || [],
          total: response.count || 0,
        },
      });
    },
    *getProjectimageTag(_, { call, put }) {
      const { list } = yield call(getProjectimageTag, _.payload);

      const tags = [];

      let branchs = list.map((item) => item.branch_name);

      branchs = [...new Set(branchs)];

      for (let i = 0; i < branchs.length; i += 1) {
        const branch = branchs[i];
        const childs = [];
        for (let j = 0; i < list.length; j += 1) {
          const item = list[j];
          if (item.branch_name === branch) {
            childs.push({ value: item.commit_id, label: item.commit_id });
          }
        }

        tags.push({ value: branch, label: branch, children: childs });
      }

      // console.log('tags==>',tags, response);

      yield put({
        type: 'save',
        payload: {
          imageTags: list,
          imageTagsOption: tags,
        },
      });
      // yield put({
      //   type: 'save',
      //   payload: {
      //     imageTags: tags
      //   },
      // });
    },

    *getAllLabels({ payload }, { call, put }) {
      const result = yield call(getAllProjectLabels, payload);
      yield put({
        type: 'save',
        payload: {
          allLabels: result || [],
        },
      });
    },
    *getImageArgTemplate({ payload: { teamId } }, { call, put }) {
      const result = yield call(getTemplates, teamId, 1, 20);
      yield put({
        type: 'save',
        payload: {
          imageArgTemplates: result.list || [],
        },
      });
    },
    *getMemberRole({ payload }, { call, put }) {
      const result = yield call(getMemberRole, payload);
      yield put({
        type: 'save',
        payload: {
          memberRole: result,
        },
      });
    },
    *getProjectClusters({ payload: { projectId } }, { call, put, select }) {
      const projectClusterMap = yield select((state) => {
        return state.application.projectClusterMap;
      });

      try {
        if (!projectClusterMap[projectId]) {
          const { list } = yield call(getProjectClusters, projectId);
          yield put({
            type: 'saveClusterMap',
            payload: {
              key: projectId,
              value: list,
            },
          });
        }
      } catch (error) {
        message.error(error.message);
      }
    },
  },
  reducers: {
    saveClusterMap(state, { payload: { key, value } }) {
      return {
        ...state,
        projectClusterMap: {
          ...state.projectClusterMap,
          [key]: value,
        },
      };
    },
    save(state, { payload }) {
      // console.log('sava::==>',state, payload);
      return { ...state, ...payload };
    },
    saveAll(state, { payload }) {
      // console.log('sava::==>',state, payload);
      return { ...state, ...payload };
    },
    clear() {
      return initialValues;
    },
  },
};
export default Model;
