import {
  queryActivities,
  queryTeam,
  queryUsedProjects,
  queryActivitiesLog,
} from './service';

const Model = {
  namespace: 'dashboard',
  state: {
    // currentUser: undefined,

    activities: [],

    teams: [],
    usedProjects: [],
    activitiesLogs: [],
  },
  effects: {
    *init(_, { put }) {
      // yield put({
      //   type: 'fetchActivitiesList',
      // });
      yield put({
        type: 'fetchTeam',
      });
      yield put({
        type: 'fetchUsedProjects',
      });
      yield put({
        type: 'fetchActivitiesLog',
      });
      // console.log('ok');
    },

    // *fetchUserCurrent(_, { call, put }) {
    //   const response = yield call(queryCurrent);
    //   yield put({
    //     type: 'save',
    //     payload: {
    //       currentUser: response,
    //     },
    //   });
    // },

    *fetchActivitiesList(_, { call, put }) {
      const response = yield call(queryActivities);
      yield put({
        type: 'save',
        payload: {
          activities: Array.isArray(response) ? response : [],
        },
      });
    },

    // *fetchChart(_, { call, put }) {
    //   const { radarData } = yield call(fakeChartData);
    //   yield put({
    //     type: 'save',
    //     payload: {
    //       radarData,
    //     },
    //   });
    // },
    *fetchTeam(_, { call, put }) {
      const { list } = yield call(queryTeam);

      yield put({
        type: 'save',
        payload: {
          teams: list,
        },
      });
    },
    *fetchUsedProjects(_, { call, put }) {
      const rep = yield call(queryUsedProjects);
      // console.log('用户使用的app==>',rep);
      yield put({
        type: 'save',
        payload: {
          usedProjects: rep,
        },
      });
    },
    *fetchActivitiesLog(_, { call, put }) {
      const { list } = yield call(queryActivitiesLog);
      //  console.log('用户使用的app==>',rep);
      yield put({
        type: 'save',
        payload: {
          activitiesLogs: list,
        },
      });
    },
  },
  reducers: {
    save(state, { payload }) {
      return { ...state, ...payload };
    },

    clear() {
      return {
        // currentUser: undefined,
        activities: [],
        teams: [],
        usedProjects: [],
      };
    },
  },
};
export default Model;
