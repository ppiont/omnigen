import { apiRequest } from '../utils/api';

export const jobIntervention = {
  pause: (jobId) => apiRequest(`/api/v1/jobs/${jobId}/pause`, { method: 'POST' }),
  
  resume: (jobId) => apiRequest(`/api/v1/jobs/${jobId}/resume`, { method: 'POST' }),
  
  modifyScene: (jobId, sceneNumber, data) => 
    apiRequest(`/api/v1/jobs/${jobId}/scenes/${sceneNumber}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    }),
    
  skipScene: (jobId, sceneNumber) => 
    apiRequest(`/api/v1/jobs/${jobId}/scenes/${sceneNumber}`, { method: 'DELETE' }),
    
  updateParameters: (jobId, params) => 
    apiRequest(`/api/v1/jobs/${jobId}/parameters`, {
      method: 'PUT',
      body: JSON.stringify(params)
    })
};
