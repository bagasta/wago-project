import api from './api';

export const generatePin = async () => {
    const response = await api.post('/auth/generate-pin');
    return response.data;
};

export const login = async (pin) => {
    // Basic Auth
    const response = await api.post('/auth/login', {}, {
        auth: {
            username: pin,
            password: ''
        }
    });
    return response.data;
};

export const logout = async () => {
    const response = await api.post('/auth/logout');
    return response.data;
};
