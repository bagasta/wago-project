import { Fragment, useEffect, useState } from 'react';
import { Dialog, Transition } from '@headlessui/react';
import { X, Activity, MessageSquare, ArrowDownLeft, ArrowUpRight, Zap, Clock, Download } from 'lucide-react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts';
import api from '../../services/api';
import { toast } from 'sonner';

export default function AnalyticsModal({ isOpen, onClose, session }) {
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);
    const [exporting, setExporting] = useState(false);

    useEffect(() => {
        if (isOpen && session) {
            fetchAnalytics();
        }
    }, [isOpen, session]);

    const fetchAnalytics = async () => {
        setLoading(true);
        try {
            const response = await api.get(`/sessions/${session.session_id}/analytics`);
            setStats(response.data);
        } catch (error) {
            console.error('Failed to fetch analytics:', error);
            if (error.response) {
                console.error('Error response:', error.response.status, error.response.data);
            }
            toast.error('Failed to load analytics');
        } finally {
            setLoading(false);
        }
    };

    const handleExportContacts = async () => {
        setExporting(true);
        try {
            const response = await api.get(`/sessions/${session.session_id}/contacts`);
            const contacts = response.data;

            if (!contacts || contacts.length === 0) {
                toast.info('No contacts found to export');
                return;
            }

            // Convert to CSV
            const headers = ['Phone Number', 'Last Active', 'Message Count'];
            const csvContent = [
                headers.join(','),
                ...contacts.map(c => [
                    c.phone_number,
                    new Date(c.last_active).toISOString(),
                    c.message_count
                ].join(','))
            ].join('\n');

            // Download
            const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
            const url = URL.createObjectURL(blob);
            const link = document.createElement('a');
            link.setAttribute('href', url);
            link.setAttribute('download', `contacts_${session.session_name}_${new Date().toISOString().split('T')[0]}.csv`);
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);

            toast.success(`Exported ${contacts.length} contacts`);
        } catch (error) {
            console.error('Failed to export contacts:', error);
            toast.error('Failed to export contacts');
        } finally {
            setExporting(false);
        }
    };

    if (!isOpen) return null;

    return (
        <Transition appear show={isOpen} as={Fragment}>
            <Dialog as="div" className="relative z-50" onClose={onClose}>
                <Transition.Child
                    as={Fragment}
                    enter="ease-out duration-300"
                    enterFrom="opacity-0"
                    enterTo="opacity-100"
                    leave="ease-in duration-200"
                    leaveFrom="opacity-100"
                    leaveTo="opacity-0"
                >
                    <div className="fixed inset-0 bg-black/80 backdrop-blur-sm" />
                </Transition.Child>

                <div className="fixed inset-0 overflow-y-auto">
                    <div className="flex min-h-full items-center justify-center p-4 text-center">
                        <Transition.Child
                            as={Fragment}
                            enter="ease-out duration-300"
                            enterFrom="opacity-0 scale-95"
                            enterTo="opacity-100 scale-100"
                            leave="ease-in duration-200"
                            leaveFrom="opacity-100 scale-100"
                            leaveTo="opacity-0 scale-95"
                        >
                            <Dialog.Panel className="w-full max-w-4xl transform overflow-hidden rounded-2xl bg-gray-900 border border-gray-800 p-6 text-left align-middle shadow-xl transition-all">
                                <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
                                    <div>
                                        <Dialog.Title as="h3" className="text-xl font-bold text-white flex items-center gap-2">
                                            <Activity className="w-5 h-5 text-blue-500" />
                                            Session Analytics
                                        </Dialog.Title>
                                        <p className="text-sm text-gray-400 mt-1">
                                            {session?.session_name} • {session?.phone_number || 'No Phone Number'}
                                        </p>
                                    </div>
                                    <div className="flex items-center gap-2 w-full sm:w-auto">
                                        <button
                                            onClick={handleExportContacts}
                                            disabled={exporting}
                                            className="flex-1 sm:flex-none justify-center p-2 rounded-lg bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 transition-colors flex items-center gap-2 text-sm font-medium"
                                            title="Export Contacts to CSV"
                                        >
                                            <Download className="w-4 h-4" />
                                            <span className="sm:inline">Export Contacts</span>
                                        </button>
                                        <button
                                            onClick={onClose}
                                            className="p-2 rounded-lg hover:bg-gray-800 text-gray-400 hover:text-white transition-colors"
                                        >
                                            <X className="w-5 h-5" />
                                        </button>
                                    </div>
                                </div>

                                {loading ? (
                                    <div className="h-64 flex items-center justify-center">
                                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
                                    </div>
                                ) : stats ? (
                                    <div className="space-y-6">
                                        {/* Key Metrics */}
                                        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                                            <div className="bg-gray-800/50 p-4 rounded-xl border border-gray-700/50">
                                                <div className="flex items-center gap-2 text-gray-400 mb-2">
                                                    <MessageSquare className="w-4 h-4" />
                                                    <span className="text-xs font-medium uppercase tracking-wider">Total Messages</span>
                                                </div>
                                                <p className="text-2xl font-bold text-white">{stats.total_messages}</p>
                                            </div>
                                            <div className="bg-gray-800/50 p-4 rounded-xl border border-gray-700/50">
                                                <div className="flex items-center gap-2 text-green-400 mb-2">
                                                    <ArrowDownLeft className="w-4 h-4" />
                                                    <span className="text-xs font-medium uppercase tracking-wider">Incoming</span>
                                                </div>
                                                <p className="text-2xl font-bold text-white">{stats.incoming_messages}</p>
                                            </div>
                                            <div className="bg-gray-800/50 p-4 rounded-xl border border-gray-700/50">
                                                <div className="flex items-center gap-2 text-blue-400 mb-2">
                                                    <ArrowUpRight className="w-4 h-4" />
                                                    <span className="text-xs font-medium uppercase tracking-wider">AI Replies</span>
                                                </div>
                                                <p className="text-2xl font-bold text-white">{stats.outgoing_messages}</p>
                                            </div>
                                            <div className="bg-gray-800/50 p-4 rounded-xl border border-gray-700/50">
                                                <div className="flex items-center gap-2 text-purple-400 mb-2">
                                                    <Zap className="w-4 h-4" />
                                                    <span className="text-xs font-medium uppercase tracking-wider">Webhook Success</span>
                                                </div>
                                                <p className="text-2xl font-bold text-white">{stats.webhook_success_rate.toFixed(1)}%</p>
                                            </div>
                                        </div>

                                        {/* Secondary Metrics */}
                                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                            <div className="bg-gray-800/50 p-4 rounded-xl border border-gray-700/50 flex justify-between items-center">
                                                <div>
                                                    <div className="flex items-center gap-2 text-gray-400 mb-1">
                                                        <Clock className="w-4 h-4" />
                                                        <span className="text-xs font-medium uppercase tracking-wider">Avg Response Time</span>
                                                    </div>
                                                    <p className="text-xl font-bold text-white">{stats.avg_response_time.toFixed(0)}ms</p>
                                                </div>
                                                <div className="h-10 w-24">
                                                    {/* Mini sparkline placeholder */}
                                                </div>
                                            </div>
                                            <div className="bg-gray-800/50 p-4 rounded-xl border border-gray-700/50 flex justify-between items-center">
                                                <div>
                                                    <div className="flex items-center gap-2 text-gray-400 mb-1">
                                                        <Activity className="w-4 h-4" />
                                                        <span className="text-xs font-medium uppercase tracking-wider">Last Active</span>
                                                    </div>
                                                    <p className="text-xl font-bold text-white">
                                                        {stats.last_active ? new Date(stats.last_active).toLocaleString() : 'Never'}
                                                    </p>
                                                </div>
                                            </div>
                                        </div>

                                        {/* Chart */}
                                        <div className="bg-gray-800/50 p-6 rounded-xl border border-gray-700/50">
                                            <h4 className="text-sm font-medium text-gray-400 mb-6">Message Activity (Last 7 Days)</h4>
                                            <div className="h-64 w-full">
                                                <ResponsiveContainer width="100%" height="100%">
                                                    <AreaChart data={stats.daily_stats}>
                                                        <defs>
                                                            <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                                                                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                                                                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                                                            </linearGradient>
                                                        </defs>
                                                        <CartesianGrid strokeDasharray="3 3" stroke="#374151" vertical={false} />
                                                        <XAxis
                                                            dataKey="date"
                                                            stroke="#9ca3af"
                                                            fontSize={12}
                                                            tickLine={false}
                                                            axisLine={false}
                                                        />
                                                        <YAxis
                                                            stroke="#9ca3af"
                                                            fontSize={12}
                                                            tickLine={false}
                                                            axisLine={false}
                                                        />
                                                        <Tooltip
                                                            contentStyle={{ backgroundColor: '#1f2937', borderColor: '#374151', color: '#fff' }}
                                                            itemStyle={{ color: '#fff' }}
                                                        />
                                                        <Area
                                                            type="monotone"
                                                            dataKey="count"
                                                            stroke="#3b82f6"
                                                            strokeWidth={2}
                                                            fillOpacity={1}
                                                            fill="url(#colorCount)"
                                                        />
                                                    </AreaChart>
                                                </ResponsiveContainer>
                                            </div>
                                        </div>
                                    </div>
                                ) : (
                                    <div className="text-center py-12 text-gray-500">
                                        No analytics data available.
                                    </div>
                                )}
                            </Dialog.Panel>
                        </Transition.Child>
                    </div>
                </div>
            </Dialog>
        </Transition>
    );
}
