import { createContext, useContext, useState, useEffect } from 'react';
import { login as loginApi, logout as logoutApi } from '../services/auth';

const AuthContext = createContext();

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const token = localStorage.getItem('token');
        const userId = localStorage.getItem('user_id');
        if (token && userId) {
            setUser({ id: userId });
        }
        setLoading(false);
    }, []);

    const login = async (pin) => {
        try {
            const response = await loginApi(pin);
            if (response.success) {
                const { token, user_id } = response.data;
                localStorage.setItem('token', token);
                localStorage.setItem('user_id', user_id);
                setUser({ id: user_id });
                return true;
            }
            return false;
        } catch (error) {
            console.error('Login failed:', error);
            throw error;
        }
    };

    const logout = async () => {
        try {
            await logoutApi();
        } catch (error) {
            console.error('Logout failed:', error);
        } finally {
            localStorage.removeItem('token');
            localStorage.removeItem('user_id');
            setUser(null);
        }
    };

    return (
        <AuthContext.Provider value={{ user, login, logout, loading }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);
