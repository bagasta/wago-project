import api from './api';

export const getSessions = async () => {
    const response = await api.get('/sessions');
    return response.data;
};

export const createSession = async (data) => {
    const response = await api.post('/sessions', data);
    return response.data;
};

export const startSession = async (id) => {
    const response = await api.post(`/sessions/${id}/start`);
    return response.data;
};

export const stopSession = async (id) => {
    const response = await api.post(`/sessions/${id}/stop`);
    return response.data;
};

export const deleteSession = async (id) => {
    const response = await api.delete(`/sessions/${id}`);
    return response.data;
};

export const updateSession = async (id, data) => {
    const response = await api.put(`/sessions/${id}`, data);
    return response.data;
};
