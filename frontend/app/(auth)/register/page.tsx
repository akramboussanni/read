'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { UserPlus, User, Mail, Lock, Check, BookOpen, Sparkles, Users, ArrowRight, ArrowLeft, GraduationCap, Shield } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { cn } from '@/lib/utils';

export default function RegisterPage() {
  const router = useRouter();
  const { register, isLoading, error, clearError, isAuthenticated } = useAuthStore();

  // Multi-step state
  const [step, setStep] = useState(1);
  const [role, setRole] = useState<'student' | 'teacher' | null>(null);
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [classCode, setClassCode] = useState('');
  const [validationError, setValidationError] = useState('');

  useEffect(() => {
    if (isAuthenticated) {
      router.push('/');
    }
  }, [isAuthenticated, router]);

  const handleNext = () => {
    setValidationError('');
    clearError();

    if (step === 1 && !role) {
      setValidationError('Veuillez choisir un rôle');
      return;
    }

    if (step === 2) {
      if (username.length < 3) {
        setValidationError("Le nom d'utilisateur doit contenir au moins 3 caractères");
        return;
      }
      if (!/^[a-zA-Z0-9_-]+$/.test(username)) {
        setValidationError("Le nom d'utilisateur ne peut contenir que des lettres, des chiffres, des traits d'union et des tirets bas");
        return;
      }
    }

    if (step === 3) {
      if (role === 'teacher' && !email) {
        setValidationError('L\'adresse email est obligatoire pour les enseignants');
        return;
      }
      if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
        setValidationError('Veuillez entrer une adresse email valide');
        return;
      }
    }

    setStep(step + 1);
  };

  const handleBack = () => {
    setStep(step - 1);
    setValidationError('');
    clearError();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();
    setValidationError('');

    if (password.length < 8) {
      setValidationError('Le mot de passe doit contenir au moins 8 caractères');
      return;
    }

    if (password !== confirmPassword) {
      setValidationError('Les mots de passe ne correspondent pas');
      return;
    }

    try {
      const confirmUrl = email ? `${window.location.origin}/confirm-email` : undefined;
      await register(username, password, email || undefined, confirmUrl, role || 'student');

      // If class code provided and student, try to join
      if (classCode && role === 'student') {
        const { classroomApi } = await import('@/lib/api/classroom');
        try {
          await classroomApi.joinClassroom(classCode);
        } catch (e) {
          console.error('Failed to auto-join class:', e);
        }
      }

      router.push('/');
    } catch (error) {
      // Error is handled by the store
    }
  };

  const totalSteps = role === 'student' ? 5 : 4;
  const progress = (step / totalSteps) * 100;

  return (
    <div className="flex items-center justify-center min-h-screen p-4 bg-background overflow-hidden relative">
      {/* Animated blob decorations */}
      <div className="blob-green top-20 right-10 animate-float opacity-30" style={{ animationDelay: '0.5s' }} />
      <div className="blob-orange top-40 left-20 animate-wobble opacity-20" style={{ animationDelay: '2s' }} />
      <div className="blob-teal bottom-32 right-1/4 animate-float opacity-20" style={{ animationDelay: '3.5s' }} />

      <div className="w-full max-w-md relative z-10">
        <Card className="fun-card border-secondary/30 p-4 mt-8 bg-white/90 backdrop-blur-md shadow-2xl">
          <CardHeader className="text-center space-y-4 pb-2">
            <div className="w-full h-2 bg-muted rounded-full overflow-hidden mb-4">
              <motion.div
                className="h-full bg-primary"
                initial={{ width: 0 }}
                animate={{ width: `${progress}%` }}
                transition={{ duration: 0.5 }}
              />
            </div>

            <AnimatePresence mode="wait">
              <motion.div
                key={step}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.3 }}
              >
                {step === 1 && (
                  <>
                    <div className="mx-auto w-20 h-20 bg-primary text-white rounded-2xl flex items-center justify-center shadow-lg border-b-6 border-primary-hover mb-4">
                      <Users className="w-10 h-10" strokeWidth={3} />
                    </div>
                    <CardTitle className="text-3xl font-black">Qui es-tu ?</CardTitle>
                    <CardDescription className="text-base font-bold text-muted-foreground">Dis-nous quel est ton rôle</CardDescription>
                  </>
                )}
                {step === 2 && (
                  <>
                    <div className="mx-auto w-20 h-20 bg-secondary text-white rounded-2xl flex items-center justify-center shadow-lg border-b-6 border-secondary-hover mb-4">
                      <User className="w-10 h-10" strokeWidth={3} />
                    </div>
                    <CardTitle className="text-3xl font-black">Ton Identité</CardTitle>
                    <CardDescription className="text-base font-bold text-muted-foreground">Comment souhaites-tu être appelé ?</CardDescription>
                  </>
                )}
                {step === 3 && (
                  <>
                    <div className="mx-auto w-20 h-20 bg-teal-500 text-white rounded-2xl flex items-center justify-center shadow-lg border-b-6 border-teal-700 mb-4">
                      <Mail className="w-10 h-10" strokeWidth={3} />
                    </div>
                    <CardTitle className="text-3xl font-black">Contact</CardTitle>
                    <CardDescription className="text-base font-bold text-muted-foreground">Une étape importante pour ton suivi</CardDescription>
                  </>
                )}
                {step === 4 && (
                  <>
                    <div className="mx-auto w-20 h-20 bg-purple-500 text-white rounded-2xl flex items-center justify-center shadow-lg border-b-6 border-purple-700 mb-4">
                      <Lock className="w-10 h-10" strokeWidth={3} />
                    </div>
                    <CardTitle className="text-3xl font-black">Sécurité</CardTitle>
                    <CardDescription className="text-base font-bold text-muted-foreground">Prêt à commencer l'aventure ?</CardDescription>
                  </>
                )}
                {step === 5 && (
                  <>
                    <div className="mx-auto w-20 h-20 bg-primary text-white rounded-2xl flex items-center justify-center shadow-lg border-b-6 border-primary-hover mb-4">
                      <Users className="w-10 h-10" strokeWidth={3} />
                    </div>
                    <CardTitle className="text-3xl font-black">Code de Classe</CardTitle>
                    <CardDescription className="text-base font-bold text-muted-foreground">Rejoins ton groupe pour commencer</CardDescription>
                  </>
                )}
              </motion.div>
            </AnimatePresence>
          </CardHeader>

          <CardContent className="mt-4">
            {(error || validationError) && (
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                className="p-3 mb-4 text-sm font-bold text-red-600 bg-red-50 border-2 border-red-200 rounded-xl"
              >
                {validationError || error}
              </motion.div>
            )}

            <AnimatePresence mode="wait">
              <motion.div
                key={step}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.2 }}
                className="space-y-4"
              >
                {step === 1 && (
                  <div className="grid grid-cols-1 gap-4">
                    <button
                      type="button"
                      onClick={() => setRole('student')}
                      className={cn(
                        "flex items-center gap-4 p-5 rounded-2xl border-4 transition-all group",
                        role === 'student' ? "border-primary bg-primary/5 scale-105" : "border-muted hover:border-primary/30"
                      )}
                    >
                      <div className={cn(
                        "w-14 h-14 rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform",
                        role === 'student' ? "bg-primary text-white shadow-lg" : "bg-muted text-muted-foreground"
                      )}>
                        <GraduationCap className="w-8 h-8" strokeWidth={2.5} />
                      </div>
                      <div className="text-left">
                        <p className="font-black text-lg">Élève</p>
                        <p className="text-sm text-muted-foreground font-semibold">Je veux apprendre et progresser</p>
                      </div>
                      {role === 'student' && <Check className="ml-auto w-6 h-6 text-primary" />}
                    </button>
                    <button
                      type="button"
                      onClick={() => setRole('teacher')}
                      className={cn(
                        "flex items-center gap-4 p-5 rounded-2xl border-4 transition-all group",
                        role === 'teacher' ? "border-secondary bg-secondary/5 scale-105" : "border-muted hover:border-secondary/30"
                      )}
                    >
                      <div className={cn(
                        "w-14 h-14 rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform",
                        role === 'teacher' ? "bg-secondary text-white shadow-lg" : "bg-muted text-muted-foreground"
                      )}>
                        <Shield className="w-8 h-8" strokeWidth={2.5} />
                      </div>
                      <div className="text-left">
                        <p className="font-black text-lg">Enseignant</p>
                        <p className="text-sm text-muted-foreground font-semibold">Je veux guider mes élèves</p>
                      </div>
                    </button>
                  </div>
                )}

                {step === 2 && (
                  <div className="space-y-2">
                    <Label htmlFor="username" className="text-base font-bold">Nom d'utilisateur</Label>
                    <Input
                      id="username"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      placeholder="Ex: Explorateur99"
                      className="h-14 rounded-2xl text-lg font-bold border-2 focus-visible:ring-primary/20"
                      autoFocus
                    />
                    <p className="text-xs text-muted-foreground font-medium pl-1">
                      Minimum 3 caractères (lettres, chiffres, _ ou -)
                    </p>
                  </div>
                )}

                {step === 3 && (
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="email" className="text-base font-bold">
                        Email {role === 'teacher' && <span className="text-red-500">*</span>}
                      </Label>
                      <Input
                        id="email"
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        placeholder="votre@email.com"
                        className="h-14 rounded-2xl text-lg font-bold border-2"
                        autoFocus
                      />
                      {role === 'teacher' ? (
                        <p className="text-xs text-red-500 font-bold px-1">
                          Obligatoire pour les enseignants (vérification requise).
                        </p>
                      ) : (
                        <p className="text-xs text-muted-foreground font-medium px-1">
                          Optionnel pour les élèves, recommandé pour récupérer son compte.
                        </p>
                      )}
                    </div>
                  </div>
                )}

                {step === 4 && (
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="password" className="text-base font-bold">Mot de passe</Label>
                      <Input
                        id="password"
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="••••••••"
                        className="h-14 rounded-2xl text-lg font-bold border-2"
                        autoFocus
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="confirmPassword" className="text-base font-bold">Confirmer le mot de passe</Label>
                      <Input
                        id="confirmPassword"
                        type="password"
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        placeholder="••••••••"
                        className="h-14 rounded-2xl text-lg font-bold border-2"
                      />
                    </div>
                  </div>
                )}

                {step === 5 && role === 'student' && (
                  <div className="space-y-4">
                    <div className="space-y-2">
                      <Label htmlFor="classCode" className="text-base font-bold">Code de classe (Optionnel)</Label>
                      <Input
                        id="classCode"
                        value={classCode}
                        onChange={(e) => setClassCode(e.target.value.toUpperCase())}
                        placeholder="Ex: ABC123"
                        className="h-14 rounded-2xl text-lg font-bold border-2 font-mono uppercase tracking-widest"
                        autoFocus
                        maxLength={10}
                      />
                      <p className="text-xs text-muted-foreground font-medium px-1">
                        Si tu as déjà un code, saisis-le ici. Sinon, tu pourras rejoindre une classe plus tard.
                      </p>
                    </div>
                  </div>
                )}
              </motion.div>
            </AnimatePresence>
          </CardContent>

          <CardFooter className="flex flex-col gap-4 pt-4">
            <div className="flex gap-3 w-full">
              {step > 1 && (
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleBack}
                  className="h-14 rounded-2xl p-4 border-b-4 border-slate-200 active:border-b-0 active:translate-y-1 transition-all"
                >
                  <ArrowLeft className="w-6 h-6" />
                </Button>
              )}

              {step < totalSteps ? (
                <Button
                  type="button"
                  onClick={handleNext}
                  className="flex-1 h-14 rounded-2xl text-lg font-black bg-primary text-white border-b-4 border-primary-hover hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all flex items-center justify-center"
                >
                  Continuer <ArrowRight className="ml-2 w-5 h-5" />
                </Button>
              ) : (
                <Button
                  type="submit"
                  onClick={handleSubmit}
                  disabled={isLoading}
                  className="flex-1 h-14 rounded-2xl text-lg font-black bg-secondary text-white border-b-4 border-secondary-hover hover:-translate-y-1 active:translate-y-1 active:border-b-0 transition-all flex items-center justify-center"
                >
                  {isLoading ? (
                    <div className="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  ) : (
                    <>C'est parti ! <Sparkles className="ml-2 w-5 h-5" /></>
                  )}
                </Button>
              )}
            </div>

            <div className="text-center pt-2">
              <Link href="/login" className="text-sm font-bold text-muted-foreground hover:text-primary transition-colors">
                J'ai déjà un compte
              </Link>
            </div>
          </CardFooter>
        </Card>

        <div className="text-center mt-8 animate-slide-up opacity-60">
          <p className="text-xs text-muted-foreground font-medium italic">
            "Lis, au nom de ton Seigneur" - Sourate Al-Alaq
          </p>
        </div>
      </div>
    </div>
  );
}
