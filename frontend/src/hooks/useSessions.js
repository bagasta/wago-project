import { useState, useEffect, useCallback } from 'react';
import { getSessions, createSession, startSession, deleteSession } from '../services/session';
import { toast } from 'sonner';

export function useSessions() {
    const [sessions, setSessions] = useState([]);
    const [isLoading, setIsLoading] = useState(true);

    const fetchSessions = useCallback(async () => {
        try {
            const response = await getSessions();
            if (response.success) {
                setSessions(response.data || []);
            }
        } catch (error) {
            toast.error('Failed to fetch sessions');
        } finally {
            setIsLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchSessions();
    }, [fetchSessions]);

    const addSession = async (data) => {
        try {
            const response = await createSession(data);
            if (response.success) {
                toast.success('Session created successfully');
                fetchSessions();
                return response.data;
            }
        } catch (error) {
            toast.error(error.response?.data?.message || 'Failed to create session');
            throw error;
        }
    };

    const removeSession = async (id) => {
        try {
            await deleteSession(id);
            toast.success('Session deleted successfully');
            setSessions(prev => prev.filter(s => s.session_id !== id));
        } catch (error) {
            toast.error('Failed to delete session');
        }
    };

    const connectSession = async (id) => {
        try {
            const response = await startSession(id);
            if (response.success) {
                return response.data;
            }
        } catch (error) {
            toast.error('Failed to start session');
            throw error;
        }
    };

    const updateSessionStatus = (id, status, data = {}) => {
        setSessions(prev => prev.map(s => {
            if (s.session_id === id) {
                return { ...s, status, ...data };
            }
            return s;
        }));
    };

    return {
        sessions,
        isLoading,
        fetchSessions,
        addSession,
        removeSession,
        connectSession,
        updateSessionStatus
    };
}
