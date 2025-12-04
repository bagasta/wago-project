import { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { useSessions } from '../hooks/useSessions';
import { useWebSocket } from '../hooks/useWebSocket';
import SessionForm from '../components/ui/SessionForm';
import QRCodeModal from '../components/ui/QRCodeModal';
import AnalyticsModal from '../components/ui/AnalyticsModal';
import { Plus, Trash2, QrCode, Smartphone, Wifi, WifiOff, Loader2, Edit2, BarChart2, MessageSquare, Power } from 'lucide-react';
import { toast } from 'sonner';

export default function Dashboard() {
    const { logout } = useAuth();
    const { sessions, isLoading, addSession, removeSession, connectSession, disconnectSession, updateSessionStatus, updateSession } = useSessions();
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);
    const [activeSessionId, setActiveSessionId] = useState(null);
    const [qrCode, setQrCode] = useState(null);
    const [isQRModalOpen, setIsQRModalOpen] = useState(false);
    const [isAnalyticsModalOpen, setIsAnalyticsModalOpen] = useState(false);
    const [selectedSession, setSelectedSession] = useState(null);
    const [editingSession, setEditingSession] = useState(null);

    const [connectingSessionId, setConnectingSessionId] = useState(null);

    // WebSocket for active session (only one at a time for now to save resources, or we can listen to all?)
    // Ideally we should have a global WS or one per session card.
    // For simplicity, let's just listen to the one active in QR modal OR all of them.
    // But `useWebSocket` hook is designed for one.
    // Let's make it so that we listen to WS when we click "Connect" or when status is 'qr'.
    // Actually, we need real-time status updates for ALL sessions.
    // But creating N websockets is bad.
    // The backend supports `/ws/sessions/{id}`.
    // Maybe we should have a single WS endpoint for user? `/ws/user`?
    // But current backend implementation is per session.
    // Let's stick to: Open WS only when viewing QR or expecting status change.
    // OR, we can just poll for status if not critical.
    // But for QR, we definitely need WS.

    // Let's use WS only for the session currently being connected/viewed.
    useWebSocket(activeSessionId, (message) => {
        if (message.type === 'qr_update') {
            setQrCode(message.data.qr_code);
            setIsQRModalOpen(true);
        } else if (message.type === 'status_update') {
            updateSessionStatus(activeSessionId, message.data.status, {
                phone_number: message.data.phone_number,
                device_info: message.data.device_info
            });
            if (message.data.status === 'connected') {
                setIsQRModalOpen(false);
                setQrCode(null);
                toast.success('WhatsApp Connected!');
                setActiveSessionId(null); // Close WS to save resources? Or keep it for messages?
            }
        }
    });

    const handleCreateSession = async (data) => {
        await addSession(data);
        setIsAddModalOpen(false);
    };

    const handleEditSession = (session) => {
        setEditingSession(session);
        setIsAddModalOpen(true);
    };

    const handleUpdateSession = async (data) => {
        await updateSession(editingSession.session_id, data);
        setIsAddModalOpen(false);
        setEditingSession(null);
    };

    const handleCloseModal = () => {
        setIsAddModalOpen(false);
        setEditingSession(null);
    };

    const handleConnect = async (session) => {
        setConnectingSessionId(session.session_id);
        setQrCode(null);
        try {
            const data = await connectSession(session.session_id);
            // Backend returns status. If 'qr', it will send QR via WS.
            // If 'connected', it returns connected.
            if (data.status === 'connected') {
                toast.success('WhatsApp Connected!');
                updateSessionStatus(session.session_id, 'connected');
                // No need to set activeSessionId (WS) if already connected
            } else {
                setActiveSessionId(session.session_id);
                setIsQRModalOpen(true);
            }
        } catch (error) {
            setActiveSessionId(null);
        } finally {
            setConnectingSessionId(null);
        }
    };

    const handleCloseQR = () => {
        setIsQRModalOpen(false);
        setActiveSessionId(null); // Disconnect WS
    };

    const handleViewAnalytics = (session) => {
        setSelectedSession(session);
        setIsAnalyticsModalOpen(true);
    };

    if (isLoading) {
        return (
            <div className="min-h-screen bg-gray-950 flex items-center justify-center">
                <Loader2 className="w-8 h-8 animate-spin text-blue-500" />
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-gray-950 text-white p-6 md:p-8">
            <div className="max-w-7xl mx-auto space-y-8">
                {/* Header */}
                <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
                        <p className="text-gray-400">Manage your WhatsApp sessions and integrations</p>
                    </div>
                    <div className="flex items-center gap-3 w-full md:w-auto">
                        <button
                            onClick={() => setIsAddModalOpen(true)}
                            className="flex-1 md:flex-none justify-center flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 rounded-lg font-medium transition-colors"
                        >
                            <Plus className="w-4 h-4" /> Add Session
                        </button>
                        <button
                            onClick={logout}
                            className="px-4 py-2 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg transition-colors"
                        >
                            Logout
                        </button>
                    </div>
                </div>

                {/* Stats Overview */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div className="p-6 bg-gray-900 rounded-xl border border-gray-800">
                        <h3 className="text-sm font-medium text-gray-400 mb-2">Total Sessions</h3>
                        <p className="text-3xl font-bold text-white">{sessions.length}</p>
                    </div>
                    <div className="p-6 bg-gray-900 rounded-xl border border-gray-800">
                        <h3 className="text-sm font-medium text-gray-400 mb-2">Active Connections</h3>
                        <p className="text-3xl font-bold text-green-400">
                            {sessions.filter(s => s.status === 'connected').length}
                        </p>
                    </div>
                </div>

                {/* Sessions Grid */}
                <div>
                    <h2 className="text-xl font-semibold mb-4">Your Sessions</h2>
                    {sessions.length === 0 ? (
                        <div className="text-center py-12 bg-gray-900/50 rounded-xl border border-gray-800 border-dashed">
                            <p className="text-gray-400">No sessions found. Create one to get started.</p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                            {sessions.map((session) => (
                                <div key={session.session_id} className="bg-gray-900 rounded-xl border border-gray-800 p-6 space-y-4 hover:border-gray-700 transition-colors">
                                    <div className="flex justify-between items-start">
                                        <div>
                                            <h3 className="font-semibold text-lg">{session.session_name}</h3>
                                            <p className="text-xs text-gray-500 font-mono mt-1 truncate max-w-[200px]" title={session.webhook_url}>
                                                {session.webhook_url}
                                            </p>
                                        </div>
                                        <div className={`px-2 py-1 rounded-full text-xs font-medium flex items-center gap-1 ${session.status === 'connected' ? 'bg-green-500/10 text-green-400' :
                                            session.status === 'qr' ? 'bg-yellow-500/10 text-yellow-400' :
                                                'bg-red-500/10 text-red-400'
                                            }`}>
                                            {session.status === 'connected' ? <Wifi className="w-3 h-3" /> : <WifiOff className="w-3 h-3" />}
                                            {session.status.toUpperCase()}
                                        </div>
                                    </div>

                                    {session.status === 'connected' && (
                                        <div className="py-3 px-4 bg-gray-800/50 rounded-lg space-y-2">
                                            <div className="flex items-center gap-2 text-sm text-gray-300">
                                                <Smartphone className="w-4 h-4 text-gray-500" />
                                                <span>{session.phone_number || 'Unknown Number'}</span>
                                            </div>
                                            {session.device_info && (
                                                <div className="text-xs text-gray-500 pl-6">
                                                    {session.device_info.device_manufacturer} {session.device_info.device_model}
                                                </div>
                                            )}
                                        </div>
                                    )}

                                    <div className="flex items-center justify-between py-3 border-t border-gray-800">
                                        <span className="text-sm text-gray-400 flex items-center gap-2">
                                            <MessageSquare className="w-4 h-4" /> Group Response
                                        </span>
                                        <button
                                            onClick={() => updateSession(session.session_id, { is_group_response_enabled: !session.is_group_response_enabled })}
                                            className={`w-10 h-5 rounded-full relative transition-colors ${session.is_group_response_enabled ? 'bg-blue-600' : 'bg-gray-700'
                                                }`}
                                            title={session.is_group_response_enabled ? "Disable Group Response" : "Enable Group Response"}
                                        >
                                            <div className={`w-3 h-3 bg-white rounded-full absolute top-1 transition-all ${session.is_group_response_enabled ? 'left-6' : 'left-1'
                                                }`} />
                                        </button>
                                    </div>

                                    <div className="flex flex-wrap gap-3 pt-2">
                                        {session.status !== 'connected' ? (
                                            <button
                                                onClick={() => handleConnect(session)}
                                                disabled={connectingSessionId === session.session_id}
                                                className="flex-1 min-w-[120px] py-2 px-3 bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                                            >
                                                {connectingSessionId === session.session_id ? (
                                                    <Loader2 className="w-4 h-4 animate-spin" />
                                                ) : (
                                                    <QrCode className="w-4 h-4" />
                                                )}
                                                Connect
                                            </button>
                                        ) : (
                                            <button
                                                onClick={() => disconnectSession(session.session_id)}
                                                className="flex-1 min-w-[120px] py-2 px-3 bg-red-600/10 hover:bg-red-600/20 text-red-400 rounded-lg text-sm font-medium transition-colors flex items-center justify-center gap-2"
                                            >
                                                <Power className="w-4 h-4" /> Stop
                                            </button>
                                        )}
                                        <div className="flex gap-3 flex-1 justify-end">
                                            <button
                                                onClick={() => handleEditSession(session)}
                                                className="py-2 px-3 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg text-sm font-medium transition-colors"
                                                title="Edit Session"
                                            >
                                                <Edit2 className="w-4 h-4" />
                                            </button>
                                            <button
                                                onClick={() => handleViewAnalytics(session)}
                                                className="py-2 px-3 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-lg text-sm font-medium transition-colors"
                                                title="View Analytics"
                                            >
                                                <BarChart2 className="w-4 h-4" />
                                            </button>
                                            <button
                                                onClick={() => removeSession(session.session_id)}
                                                className="py-2 px-3 bg-red-500/10 hover:bg-red-500/20 text-red-400 rounded-lg text-sm font-medium transition-colors"
                                                title="Delete Session"
                                            >
                                                <Trash2 className="w-4 h-4" />
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>

            <SessionForm
                isOpen={isAddModalOpen}
                onClose={handleCloseModal}
                onSubmit={editingSession ? handleUpdateSession : handleCreateSession}
                initialData={editingSession}
            />

            <QRCodeModal
                isOpen={isQRModalOpen}
                onClose={handleCloseQR}
                qrCode={qrCode}
                status="qr"
            />

            <AnalyticsModal
                isOpen={isAnalyticsModalOpen}
                onClose={() => setIsAnalyticsModalOpen(false)}
                session={selectedSession}
            />
        </div>
    );
}
