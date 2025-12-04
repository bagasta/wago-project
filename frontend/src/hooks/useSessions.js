import { useState, useEffect, useCallback } from 'react';
import { getSessions, createSession, startSession, stopSession, deleteSession, updateSession as updateSessionService } from '../services/session';
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

    const disconnectSession = async (id) => {
        try {
            const response = await stopSession(id);
            if (response.success) {
                toast.success('Session stopped');
                updateSessionStatus(id, 'disconnected');
                return response.data;
            }
        } catch (error) {
            toast.error('Failed to stop session');
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

    const updateSession = async (id, data) => {
        try {
            const response = await updateSessionService(id, data);
            if (response.success) {
                toast.success('Session updated successfully');
                setSessions(prev => prev.map(s => {
                    if (s.session_id === id) {
                        return { ...s, ...response.data };
                    }
                    return s;
                }));
                return response.data;
            }
        } catch (error) {
            toast.error(error.response?.data?.message || 'Failed to update session');
            throw error;
        }
    };

    return {
        sessions,
        isLoading,
        fetchSessions,
        addSession,
        removeSession,
        connectSession,
        disconnectSession,
        updateSessionStatus,
        updateSession
    };
}
