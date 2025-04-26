import request from '@/utils/request';

class ServiceManager {
  async get(url, options = {}) {
    return this.generateRequest(url, { ...options, method: 'GET' });
  }

  async patch(url, options = {}) {
    return this.generateRequest(url, { ...options, method: 'PATCH' });
  }

  async post(url, options = {}) {
    return this.generateRequest(url, { ...options, method: 'POST' });
  }

  async put(url, options = {}) {
    return this.generateRequest(url, { ...options, method: 'PUT' });
  }

  async delete(url, options = {}) {
    return this.generateRequest(url, { ...options, method: 'DELETE' });
  }
  /* eslint-disable class-methods-use-this */

  async generateRequest(url, options = {}) {
    try {
      const response = await request(url, options);

      if (response.errcode !== undefined) {
        if (response.errcode === 0) {
          return response.data;
        }

        const error = new Error(response.errmsg);
        error.code = response.errcode;
        throw error;
      }
      return response.data || response;
    } catch (error) {
      error.data = {
        ...(error.data || {}),
        url,
        method: (options.method || 'GET').toLowerCase(),
        reason: 'api',
      };
      throw error;
    }
  }
}
export default new ServiceManager();
