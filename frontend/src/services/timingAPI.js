import { apiRequest } from '../utils/api';

export const timingAPI = {
  updateSceneTiming: (jobId, sceneNumber, data) =>
    apiRequest(`/api/v1/jobs/${jobId}/scenes/${sceneNumber}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  regenerateScene: (jobId, sceneNumber) =>
    apiRequest(`/api/v1/jobs/${jobId}/scenes/${sceneNumber}/regenerate`, {
      method: 'POST',
    }),
};
