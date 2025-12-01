import { useState } from 'react';
import { Link } from 'react-router-dom';
import { generatePin } from '../services/auth';
import { toast } from 'sonner';
import { Key, Copy, Check, ArrowLeft, Loader2 } from 'lucide-react';

export default function GeneratePin() {
    const [generatedData, setGeneratedData] = useState(null);
    const [isLoading, setIsLoading] = useState(false);
    const [copied, setCopied] = useState(false);

    const handleGenerate = async () => {
        setIsLoading(true);
        try {
            const response = await generatePin();
            if (response.success) {
                setGeneratedData(response.data);
                toast.success('PIN generated successfully');
            }
        } catch (error) {
            toast.error('Failed to generate PIN');
        } finally {
            setIsLoading(false);
        }
    };

    const copyToClipboard = () => {
        if (generatedData?.pin) {
            navigator.clipboard.writeText(generatedData.pin);
            setCopied(true);
            toast.success('PIN copied to clipboard');
            setTimeout(() => setCopied(false), 2000);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-950 text-white p-4">
            <div className="w-full max-w-md space-y-8">
                <div className="text-center space-y-2">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-purple-500/10 mb-4">
                        <Key className="w-8 h-8 text-purple-500" />
                    </div>
                    <h2 className="text-3xl font-bold tracking-tight">Generate New PIN</h2>
                    <p className="text-gray-400">Create a new secure access PIN for your account</p>
                </div>

                <div className="mt-8 space-y-6">
                    {generatedData ? (
                        <div className="bg-gray-900 border border-gray-800 rounded-xl p-6 space-y-4 animate-in fade-in slide-in-from-bottom-4">
                            <div className="text-center space-y-1">
                                <p className="text-sm text-gray-400">Your New PIN</p>
                                <div className="flex items-center justify-center gap-3">
                                    <span className="text-4xl font-mono font-bold tracking-wider text-purple-400">
                                        {generatedData.pin}
                                    </span>
                                    <button
                                        onClick={copyToClipboard}
                                        className="p-2 hover:bg-gray-800 rounded-lg transition-colors text-gray-400 hover:text-white"
                                        title="Copy PIN"
                                    >
                                        {copied ? <Check className="w-5 h-5 text-green-500" /> : <Copy className="w-5 h-5" />}
                                    </button>
                                </div>
                            </div>

                            <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-lg p-3 text-sm text-yellow-200 text-center">
                                ⚠️ Please save this PIN immediately. You won't be able to see it again.
                            </div>

                            <Link
                                to="/login"
                                className="block w-full py-3 px-4 bg-purple-600 hover:bg-purple-700 text-white rounded-lg font-medium text-center transition-colors"
                            >
                                Go to Login
                            </Link>
                        </div>
                    ) : (
                        <button
                            onClick={handleGenerate}
                            disabled={isLoading}
                            className="w-full flex items-center justify-center py-4 px-4 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white rounded-xl font-bold text-lg transition-all transform hover:scale-[1.02] active:scale-[0.98] shadow-lg shadow-purple-500/20"
                        >
                            {isLoading ? (
                                <Loader2 className="w-6 h-6 animate-spin" />
                            ) : (
                                "Generate Secure PIN"
                            )}
                        </button>
                    )}

                    <div className="text-center">
                        <Link to="/login" className="inline-flex items-center text-sm text-gray-400 hover:text-white transition-colors">
                            <ArrowLeft className="w-4 h-4 mr-2" /> Back to Login
                        </Link>
                    </div>
                </div>
            </div>
        </div>
    );
}
