'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';

import { Suspense } from 'react';

function ResetPasswordContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [token, setToken] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [validationError, setValidationError] = useState('');
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const tokenParam = searchParams.get('token');
    if (tokenParam) {
      setToken(tokenParam);
    } else {
      setError('Jeton de réinitialisation invalide ou manquant');
    }
  }, [searchParams]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setValidationError('');
    setIsLoading(true);

    if (password.length < 8) {
      setValidationError('Le mot de passe doit contenir au moins 8 caractères');
      setIsLoading(false);
      return;
    }

    if (password !== confirmPassword) {
      setValidationError('Les mots de passe ne correspondent pas');
      setIsLoading(false);
      return;
    }

    if (!token) {
      setError('Jeton de réinitialisation invalide ou manquant');
      setIsLoading(false);
      return;
    }

    try {
      await authApi.resetPassword({ token, new_password: password });
      setSuccess(true);
      setTimeout(() => {
        router.push('/login');
      }, 3000);
    } catch (err: any) {
      setError(err.response?.data?.message || 'Échec de la réinitialisation du mot de passe');
    } finally {
      setIsLoading(false);
    }
  };

  if (success) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Réinitialisation Réussie</CardTitle>
            <CardDescription>
              Votre mot de passe a été changé
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-600">
              Redirection vers la page de connexion...
            </p>
          </CardContent>
          <CardFooter>
            <Link href="/login" className="w-full">
              <Button className="w-full">
                Aller à la Connexion
              </Button>
            </Link>
          </CardFooter>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center min-h-screen p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Réinitialiser le Mot de Passe</CardTitle>
          <CardDescription>
            Entrez votre nouveau mot de passe
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
            {(error || validationError) && (
              <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">
                {validationError || error}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="password">Nouveau Mot de Passe</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
                disabled={isLoading || !token}
                minLength={8}
              />
              <p className="text-xs text-gray-500">
                Au moins 8 caractères
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="confirmPassword">Confirmez le Nouveau Mot de Passe</Label>
              <Input
                id="confirmPassword"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                placeholder="••••••••"
                required
                disabled={isLoading || !token}
                minLength={8}
              />
            </div>
          </CardContent>
          <CardFooter className="flex flex-col space-y-4">
            <Button
              type="submit"
              className="w-full"
              disabled={isLoading || !token}
            >
              {isLoading ? 'Réinitialisation...' : 'Réinitialiser le Mot de Passe'}
            </Button>
            <Link href="/login" className="text-sm text-center text-blue-600 hover:underline">
              Retour à la Connexion
            </Link>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Loading...</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    }>
      <ResetPasswordContent />
    </Suspense>
  );
}
