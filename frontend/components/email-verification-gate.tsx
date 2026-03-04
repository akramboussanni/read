'use client';

import React from 'react';
import { useAuthStore } from '@/lib/store/auth-store';
import { usePathname } from 'next/navigation';
import { motion } from 'framer-motion';
import { Mail, CheckCircle, RefreshCw, LogOut } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { authApi } from '@/lib/api/auth';

export function EmailVerificationGate({ children }: { children: React.ReactNode }) {
    const { user, logout } = useAuthStore();
    const pathname = usePathname();
    const [resending, setResending] = React.useState(false);
    const [sent, setSent] = React.useState(false);
    const [checking, setChecking] = React.useState(false);

    // Debugging current user state
    React.useEffect(() => {
        if (user) {
            console.log('Verification Gate Check:',
                { email: user.email, role: user.role, confirmed: user.email_confirmed }
            );
        }
    }, [user]);

    // Auto-poll for verification status every 5 seconds if unverified
    React.useEffect(() => {
        let interval: NodeJS.Timeout;

        if (user && user.role === 'teacher' && !user.email_confirmed) {
            interval = setInterval(async () => {
                try {
                    const { checkAuth } = useAuthStore.getState();
                    await checkAuth();
                } catch (e) {
                    console.error('Auto-poll check failed', e);
                }
            }, 5000);
        }

        return () => {
            if (interval) clearInterval(interval);
        };
    }, [user]);

    // If user is not logged in or is not a teacher, allow access
    if (!user || user.role !== 'teacher') {
        return <>{children}</>;
    }

    // Allow access to specific paths even if unverified
    const allowedPaths = ['/confirm-email', '/profile'];
    if (allowedPaths.some(path => pathname?.startsWith(path))) {
        return <>{children}</>;
    }

    // If teacher's email is confirmed, allow access
    if (user.email_confirmed === true) {
        return <>{children}</>;
    }

    const handleResend = async () => {
        if (!user.email) return;
        setResending(true);
        try {
            // Assuming ResendEmailConfirmation endpoint exists
            await authApi.resendConfirmation({ email: user.email });
            setSent(true);
            setTimeout(() => setSent(false), 5000);
        } catch (err) {
            console.error('Failed to resend confirmation:', err);
        } finally {
            setResending(false);
        }
    };

    const handleManualCheck = async () => {
        setChecking(true);
        try {
            const { checkAuth } = useAuthStore.getState();
            await checkAuth();
        } catch (err) {
            console.error('Failed to check auth status:', err);
        } finally {
            setChecking(false);
        }
    };



    return (
        <div className="fixed inset-0 z-[100] bg-background/80 backdrop-blur-xl flex items-center justify-center p-4">
            <motion.div
                initial={{ opacity: 0, scale: 0.9, y: 20 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                className="bg-white rounded-[2rem] border-4 border-primary shadow-2xl p-8 max-w-lg w-full text-center relative overflow-hidden"
            >
                <div className="blob-green-header absolute -top-40 -right-40 opacity-20 pointer-events-none" />
                <div className="blob-orange-header absolute -bottom-40 -left-20 opacity-20 pointer-events-none" />

                <div className="relative z-10 space-y-8">
                    <div className="mx-auto w-24 h-24 bg-primary text-white rounded-3xl flex items-center justify-center shadow-lg border-b-8 border-primary-hover animate-bounce-slow">
                        <Mail className="w-12 h-12" strokeWidth={3} />
                    </div>

                    <div className="space-y-4">
                        <h1 className="text-4xl font-black tracking-tight text-slate-800">Vérifie tes emails !</h1>
                        <p className="text-slate-600 font-bold text-lg leading-relaxed">
                            En tant qu'enseignant, tu dois <span className="text-primary underline decoration-4 decoration-primary/30">valider ton adresse email</span> pour accéder à tes classes.
                        </p>
                    </div>

                    <div className="bg-slate-50 rounded-2xl p-5 border-2 border-slate-100 flex items-center gap-4 text-left">
                        <div className="w-12 h-12 bg-white rounded-xl flex items-center justify-center text-primary shadow-sm border-2 border-slate-200 shrink-0">
                            <Mail className="w-6 h-6" />
                        </div>
                        <div>
                            <p className="text-xs font-black text-slate-400 uppercase tracking-widest leading-none mb-1">Email enregistré</p>
                            <p className="text-xl font-black text-slate-700">{user.email}</p>
                        </div>
                    </div>

                    <div className="grid grid-cols-1 gap-4">
                        <Button
                            onClick={handleManualCheck}
                            disabled={checking}
                            className="h-16 rounded-2xl text-xl font-black bg-secondary text-white border-b-6 border-orange-600 hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all flex items-center justify-center gap-2"
                        >
                            {checking ? <RefreshCw className="animate-spin" /> : <CheckCircle />}
                            J'ai vérifié mon email !
                        </Button>

                        <div className="flex gap-3">
                            <Button
                                variant="outline"
                                onClick={handleResend}
                                disabled={resending || sent}
                                className="flex-1 h-16 rounded-2xl text-lg font-black border-4 border-slate-200 active:border-b-0 active:translate-y-1 transition-all"
                            >
                                {resending ? 'Envoi...' : sent ? 'C\'est envoyé ✓' : 'Renvoyer le lien'}
                            </Button>
                            <Button
                                variant="ghost"
                                onClick={() => logout()}
                                className="h-16 px-6 rounded-2xl text-lg font-black text-slate-400 hover:text-red-500 hover:bg-red-50 transition-all"
                            >
                                <LogOut className="w-6 h-6" />
                            </Button>
                        </div>
                    </div>

                    <p className="text-sm font-bold text-slate-400 italic">
                        Check tes spams si tu ne trouves rien, petit explorateur !
                    </p>
                </div>
            </motion.div>
        </div>
    );
}
