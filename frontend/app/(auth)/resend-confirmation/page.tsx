'use client';

import { useState } from 'react';
import Link from 'next/link';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';

export default function ResendConfirmationPage() {
  const [email, setEmail] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      const url = `${window.location.origin}/confirm-email`;
      await authApi.resendConfirmation({ email, url });
      setSuccess(true);
    } catch (err: any) {
      setError(err.response?.data?.message || "Échec de l'envoi de l'email de confirmation");
    } finally {
      setIsLoading(false);
    }
  };

  if (success) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Vérifiez vos Emails</CardTitle>
            <CardDescription>
              L'email de confirmation a été envoyé
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-gray-600">
              Si un compte non confirmé existe avec <strong>{email}</strong>, vous recevrez un
              lien de confirmation sous peu. Le lien expirera dans 24 heures.
            </p>
            <p className="text-sm text-gray-600">
              Vous n'avez pas reçu l'email ? Vérifiez vos spams.
            </p>
          </CardContent>
          <CardFooter>
            <Link href="/login" className="w-full">
              <Button variant="outline" className="w-full">
                Retour à la Connexion
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
          <CardTitle>Renvoyer la Confirmation</CardTitle>
          <CardDescription>
            Entrez votre email pour recevoir un nouveau lien de confirmation
          </CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
            {error && (
              <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">
                {error}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="email">Adresse Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="nom@exemple.com"
                required
                disabled={isLoading}
              />
            </div>
          </CardContent>
          <CardFooter className="flex flex-col space-y-4">
            <Button
              type="submit"
              className="w-full"
              disabled={isLoading}
            >
              {isLoading ? 'Envoi...' : 'Renvoyer la Confirmation'}
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
