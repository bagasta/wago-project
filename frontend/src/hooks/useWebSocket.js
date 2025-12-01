import { useEffect, useRef } from 'react';

export function useWebSocket(sessionId, onMessage) {
    const ws = useRef(null);
    const token = localStorage.getItem('token');

    useEffect(() => {
        if (!sessionId || !token) return;

        // Close existing connection
        if (ws.current) {
            ws.current.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host; // This handles port automatically
        // In dev, vite proxy handles /ws -> localhost:8080
        // But for WS, we need to be careful. Vite proxy supports WS.
        // So `ws://${host}/ws/sessions/${sessionId}` should work.

        const wsUrl = `${protocol}//${host}/ws/sessions/${sessionId}?token=${token}`;

        ws.current = new WebSocket(wsUrl);

        ws.current.onopen = () => {
            console.log(`WS Connected: ${sessionId}`);
        };

        ws.current.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                if (onMessage) {
                    onMessage(data);
                }
            } catch (error) {
                console.error('WS Parse Error:', error);
            }
        };

        ws.current.onerror = (error) => {
            console.error('WS Error:', error);
        };

        ws.current.onclose = () => {
            console.log(`WS Disconnected: ${sessionId}`);
        };

        return () => {
            if (ws.current) {
                ws.current.close();
            }
        };
    }, [sessionId, token]); // Re-connect if sessionId changes

    return ws.current;
}
