'use client';

import { useState, useEffect } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';

import { Suspense } from 'react';

function ConfirmEmailContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    const confirmEmail = async () => {
      const token = searchParams.get('token');

      if (!token) {
        setError('Jeton de confirmation invalide ou manquant');
        setIsLoading(false);
        return;
      }

      try {
        await authApi.confirmEmail({ token });
        setSuccess(true);
      } catch (err: any) {
        const errorMsg = err.response?.data?.message || "Échec de la confirmation de l'email";
        setError(errorMsg);
      } finally {
        setIsLoading(false);
      }
    };

    confirmEmail();
  }, [searchParams, router]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Confirmation de l'Email</CardTitle>
            <CardDescription>
              Veuillez patienter pendant que nous vérifions votre email...
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (success) {
    return (
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Email Confirmé !</CardTitle>
            <CardDescription>
              Votre email a été vérifié avec succès
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-gray-600">
              Votre email est maintenant vérifié. Vous pouvez maintenant vous connecter et utiliser toutes les fonctionnalités.
            </p>
            <p className="text-sm text-gray-600 mt-2">
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
          <CardTitle>Échec de la Confirmation</CardTitle>
          <CardDescription>
            Impossible de vérifier votre email
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="p-3 text-sm text-red-600 bg-red-50 rounded-md">
            {error}
          </div>
          <p className="text-sm text-gray-600">
            {error.includes('expired')
              ? 'Votre lien de confirmation a expiré. Veuillez en demander un nouveau.'
              : 'Le lien de confirmation est peut-être invalide ou déjà utilisé.'}
          </p>
        </CardContent>
        <CardFooter className="flex flex-col space-y-2">
          <Link href="/resend-confirmation" className="w-full">
            <Button className="w-full">
              Renvoyer l'Email de Confirmation
            </Button>
          </Link>
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

export default function ConfirmEmailPage() {
  return (
    <Suspense fallback={
      <div className="flex items-center justify-center min-h-screen p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle>Chargement...</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex justify-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-gray-900"></div>
            </div>
          </CardContent>
        </Card>
      </div>
    }>
      <ConfirmEmailContent />
    </Suspense>
  );
}
