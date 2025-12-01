import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { toast } from 'sonner';
import { Lock, ArrowRight, Loader2 } from 'lucide-react';

export default function Login() {
    const [pin, setPin] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (pin.length !== 6) {
            toast.error('PIN must be 6 characters');
            return;
        }

        setIsLoading(true);
        try {
            await login(pin);
            toast.success('Login successful');
            navigate('/dashboard');
        } catch (error) {
            toast.error(error.response?.data?.message || 'Login failed');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-950 text-white p-4">
            <div className="w-full max-w-md space-y-8">
                <div className="text-center space-y-2">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-blue-500/10 mb-4">
                        <Lock className="w-8 h-8 text-blue-500" />
                    </div>
                    <h2 className="text-3xl font-bold tracking-tight">Welcome Back</h2>
                    <p className="text-gray-400">Enter your 6-digit PIN to access your dashboard</p>
                </div>

                <form onSubmit={handleSubmit} className="mt-8 space-y-6">
                    <div className="space-y-2">
                        <label htmlFor="pin" className="text-sm font-medium text-gray-300">
                            Security PIN
                        </label>
                        <input
                            id="pin"
                            name="pin"
                            type="password"
                            maxLength={6}
                            required
                            className="w-full px-4 py-3 bg-gray-900 border border-gray-800 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all text-center text-2xl tracking-[0.5em] font-mono placeholder:tracking-normal"
                            placeholder="••••••"
                            value={pin}
                            onChange={(e) => setPin(e.target.value.toUpperCase())}
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="w-full flex items-center justify-center py-3 px-4 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {isLoading ? (
                            <Loader2 className="w-5 h-5 animate-spin" />
                        ) : (
                            <>
                                Access Dashboard <ArrowRight className="ml-2 w-5 h-5" />
                            </>
                        )}
                    </button>
                </form>

                <div className="text-center">
                    <Link to="/generate-pin" className="text-sm text-gray-400 hover:text-blue-400 transition-colors">
                        Don't have a PIN? Generate one here
                    </Link>
                </div>
            </div>
        </div>
    );
}
