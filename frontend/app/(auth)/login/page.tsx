'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/lib/store/auth-store';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { LogIn, User, Lock, BookOpen, Star } from 'lucide-react';

export default function LoginPage() {
  const router = useRouter();
  const { login, isLoading, error, clearError, isAuthenticated } = useAuthStore();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');

  useEffect(() => {
    if (isAuthenticated) {
      router.push('/');
    }
  }, [isAuthenticated, router]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();

    try {
      await login(username, password);
      router.push('/');
    } catch (error) {
      // Error is handled by the store
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen p-4 bg-background overflow-hidden relative">
      {/* Animated blob decorations */}
      <div className="blob-green top-20 left-10 animate-wobble" style={{ animationDelay: '0s' }} />
      <div className="blob-orange top-40 right-20 animate-float" style={{ animationDelay: '1.5s' }} />
      <div className="blob-teal bottom-32 left-1/4 animate-wobble" style={{ animationDelay: '3s' }} />

      <div className="w-full max-w-md relative z-10">

        <Card className="fun-card border-primary/30 p-4 shadow-2xl bg-white/80 backdrop-blur-sm">
          <CardHeader className="text-center space-y-4">
            <div className="mx-auto w-24 h-24 bg-primary text-white rounded-3xl flex items-center justify-center animate-bounce-gentle shadow-xl border-b-8 border-primary-hover mb-6">
              <LogIn className="w-12 h-12" strokeWidth={3} />
            </div>
            <div>
              <CardTitle className="text-3xl font-black text-foreground mb-2">
                Bon Retour !
              </CardTitle>
              <CardDescription className="text-base font-bold text-muted-foreground">
                Connecte-toi pour continuer de jouer
              </CardDescription>
            </div>
          </CardHeader>
          <form onSubmit={handleSubmit}>
            <CardContent className="space-y-5 animated-slide-up">
              {error && (
                <div className="p-4 text-sm font-bold text-red-600 bg-red-50 border-2 border-red-200 rounded-lg animate-bounce-gentle">
                  {error}
                </div>
              )}
              <div className="space-y-2">
                <Label htmlFor="username" className="flex items-center gap-2 text-base font-bold">
                  <User className="w-5 h-5 text-primary" />
                  Nom d'utilisateur
                </Label>
                <Input
                  id="username"
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Entrez votre nom d'utilisateur"
                  required
                  disabled={isLoading}
                  className="h-12 text-base"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password" className="flex items-center gap-2 text-base font-bold">
                  <Lock className="w-5 h-5 text-primary" />
                  Mot de passe
                </Label>
                <Input
                  id="password"
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Entrez votre mot de passe"
                  required
                  disabled={isLoading}
                  className="h-12 text-base"
                />
              </div>
              <div className="flex justify-end">
                <Link
                  href="/forgot-password"
                  className="text-sm font-bold text-primary hover:text-primary/80 transition-all hover:translate-x-1 inline-block"
                >
                  Mot de passe oublié ? →
                </Link>
              </div>
            </CardContent>
            <CardFooter className="flex flex-col space-y-5 pt-2">
              <Button
                type="submit"
                variant="secondary"
                className="w-full text-xl h-16 rounded-2xl shadow-xl hover:-translate-y-1 active:translate-y-0 active:border-b-0 transition-all font-black group"
                disabled={isLoading}
                size="lg"
              >
                {isLoading ? (
                  <span className="flex items-center gap-2">
                    <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    Connexion...
                  </span>
                ) : (
                  <span className="flex items-center gap-2">
                    Se Connecter
                    <LogIn className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
                  </span>
                )}
              </Button>
              <div className="relative w-full">
                <div className="absolute inset-0 flex items-center">
                  <div className="w-full border-t border-muted-foreground/20"></div>
                </div>
                <div className="relative flex justify-center text-sm">
                  <span className="px-4 bg-card text-muted-foreground font-semibold">ou</span>
                </div>
              </div>
              <div className="text-center space-y-2">
                <p className="text-sm text-muted-foreground">
                  Pas encore de compte ?
                </p>
                <Link href="/register">
                  <Button variant="outline" size="lg" className="w-full border-2 border-primary/30 hover:border-primary group">
                    <Star className="w-5 h-5 mr-2 group-hover:rotate-180 transition-transform duration-500" />
                    Créer un compte
                  </Button>
                </Link>
              </div>
            </CardFooter>
          </form>
        </Card>

        {/* Bottom decoration */}
        <div className="text-center mt-6 animate-slide-up">
          <p className="text-sm text-muted-foreground">
            "En vérité, c'est Nous qui avons fait descendre le Coran" - 15:9
          </p>
        </div>
      </div>
    </div>
  );
}
