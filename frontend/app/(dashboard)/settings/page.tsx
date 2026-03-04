'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/lib/store/auth-store';
import { ProtectedRoute } from '@/components/protected-route';
import { authApi } from '@/lib/api/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import Link from 'next/link';
import { motion, AnimatePresence } from 'framer-motion';
import { Shield, User, Bell, Palette, Lock, Mail, ChevronRight } from 'lucide-react';
import { ThemeToggle } from '@/components/theme-toggle';

export default function SettingsPage() {
  const { user, checkAuth } = useAuthStore();
  const [activeTab, setActiveTab] = useState('security');

  // Form States
  const [email, setEmail] = useState('');
  const [emailPassword, setEmailPassword] = useState('');
  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  // Loading & Error States
  const [isLoadingEmail, setIsLoadingEmail] = useState(false);
  const [isLoadingPassword, setIsLoadingPassword] = useState(false);
  const [emailError, setEmailError] = useState('');
  const [emailSuccess, setEmailSuccess] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const [passwordSuccess, setPasswordSuccess] = useState('');

  // Sync email state with user data
  useEffect(() => {
    if (user?.email) {
      setEmail(user.email);
    }
  }, [user?.email]);

  const handleEmailUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setEmailError('');
    setEmailSuccess('');
    setIsLoadingEmail(true);

    try {
      const url = `${window.location.origin}/confirm-email`;
      const payload: any = { email, url };

      if (user?.email && user?.email_confirmed) {
        if (!emailPassword) {
          setEmailError('Mot de passe requis pour changer un email vérifié');
          setIsLoadingEmail(false);
          return;
        }
        payload.password = emailPassword;
      }

      await authApi.addEmail(payload);
      setEmailSuccess('Email de vérification envoyé ! Vérifiez votre boîte de réception.');
      setEmailPassword('');
      await checkAuth();
    } catch (err: any) {
      setEmailError(err.response?.data?.message || "Échec de la mise à jour de l'email");
    } finally {
      setIsLoadingEmail(false);
    }
  };

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    setPasswordError('');
    setPasswordSuccess('');

    if (newPassword.length < 8) {
      setPasswordError('Le mot de passe doit contenir au moins 8 caractères');
      return;
    }

    if (newPassword !== confirmPassword) {
      setPasswordError('Les mots de passe ne correspondent pas');
      return;
    }

    setIsLoadingPassword(true);

    try {
      await authApi.changePassword({
        old_password: currentPassword,
        new_password: newPassword,
      });
      setPasswordSuccess('Mot de passe changé avec succès');
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
    } catch (err: any) {
      setPasswordError(err.response?.data?.message || 'Échec du changement de mot de passe');
    } finally {
      setIsLoadingPassword(false);
    }
  };

  const tabs = [
    { id: 'general', label: 'Général', icon: User, color: 'text-blue-500', bg: 'bg-blue-500/10' },
    { id: 'security', label: 'Sécurité', icon: Shield, color: 'text-red-500', bg: 'bg-red-500/10' },
    { id: 'appearance', label: 'Apparence', icon: Palette, color: 'text-purple-500', bg: 'bg-purple-500/10' },
    { id: 'notifications', label: 'Notifications', icon: Bell, color: 'text-yellow-500', bg: 'bg-yellow-500/10' },
  ];

  return (
    <ProtectedRoute>
      <div className="container mx-auto p-4 pt-24 max-w-6xl">
        <div className="mb-8">
          <Link href="/profile" className="text-muted-foreground hover:text-primary mb-2 inline-flex items-center text-sm font-medium transition-colors">
            <ChevronRight className="rotate-180 w-4 h-4 mr-1" />
            Retour au Profil
          </Link>
          <h1 className="text-4xl font-black bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent">
            Paramètres
          </h1>
          <p className="text-muted-foreground mt-2">Gérez les préférences de votre compte et la sécurité.</p>
        </div>

        <div className="flex flex-col lg:flex-row gap-8">
          {/* Sidebar Tabs */}
          <div className="w-full lg:w-64 flex-shrink-0 space-y-2">
            {tabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-2xl transition-all duration-200 font-bold ${activeTab === tab.id
                  ? 'bg-white dark:bg-card shadow-lg scale-105 border-2 border-primary/10'
                  : 'hover:bg-white/50 dark:hover:bg-white/5 text-muted-foreground'
                  }`}
              >
                <div className={`p-2 rounded-xl ${tab.bg} ${tab.color}`}>
                  <tab.icon className="w-5 h-5" />
                </div>
                <span>{tab.label}</span>
                {activeTab === tab.id && (
                  <motion.div
                    layoutId="active-pill"
                    className="ml-auto w-1.5 h-1.5 rounded-full bg-primary"
                  />
                )}
              </button>
            ))}
          </div>

          {/* Content Area */}
          <div className="flex-1">
            <AnimatePresence mode="wait">
              <motion.div
                key={activeTab}
                initial={{ opacity: 0, x: 20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: -20 }}
                transition={{ duration: 0.2 }}
              >
                {activeTab === 'security' && (
                  <div className="space-y-6">
                    {/* Email Card */}
                    <Card className="border-none shadow-xl bg-white/50 dark:bg-card/50 backdrop-blur-sm overflow-hidden">
                      <div className="h-2 bg-gradient-to-r from-blue-400 to-blue-600" />
                      <CardHeader>
                        <div className="flex items-center gap-3 mb-2">
                          <div className="p-2 bg-blue-100 dark:bg-blue-900/30 text-blue-600 rounded-lg">
                            <Mail className="w-6 h-6" />
                          </div>
                          <CardTitle>Adresse Email</CardTitle>
                        </div>
                        <CardDescription>
                          {user?.email
                            ? user.email_confirmed
                              ? 'Votre adresse email vérifiable pour la récupération de compte.'
                              : 'Veuillez vérifier votre adresse email.'
                            : 'Ajoutez une adresse email pour sécuriser votre compte.'}
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <form onSubmit={handleEmailUpdate} className="space-y-4">
                          {emailError && (
                            <div className="p-4 text-sm text-red-600 bg-red-50 dark:bg-red-900/10 border border-red-100 dark:border-red-900/20 rounded-xl">
                              {emailError}
                            </div>
                          )}
                          {emailSuccess && (
                            <div className="p-4 text-sm text-green-600 bg-green-50 dark:bg-green-900/10 border border-green-100 dark:border-green-900/20 rounded-xl">
                              {emailSuccess}
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
                              disabled={isLoadingEmail}
                              className="bg-white dark:bg-black/20"
                            />
                          </div>

                          {/* If verified, ask for password */}
                          {user?.email && user.email_confirmed && (
                            <div className="space-y-2 pt-2">
                              <Label htmlFor="emailPassword">Mot de Passe Actuel</Label>
                              <Input
                                id="emailPassword"
                                type="password"
                                value={emailPassword}
                                onChange={(e) => setEmailPassword(e.target.value)}
                                placeholder="Requis pour vérifier les changements"
                                required
                                disabled={isLoadingEmail}
                                className="bg-white dark:bg-black/20"
                              />
                            </div>
                          )}

                          <div className="pt-2">
                            <Button
                              type="submit"
                              disabled={isLoadingEmail || email === user?.email}
                              className="bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700 text-white border-0"
                            >
                              {isLoadingEmail ? 'Mise à jour...' : "Mettre à jour l'Email"}
                            </Button>
                          </div>
                        </form>
                      </CardContent>
                    </Card>

                    {/* Password Card */}
                    <Card className="border-none shadow-xl bg-white/50 dark:bg-card/50 backdrop-blur-sm overflow-hidden">
                      <div className="h-2 bg-gradient-to-r from-red-400 to-red-600" />
                      <CardHeader>
                        <div className="flex items-center gap-3 mb-2">
                          <div className="p-2 bg-red-100 dark:bg-red-900/30 text-red-600 rounded-lg">
                            <Lock className="w-6 h-6" />
                          </div>
                          <CardTitle>Changer le Mot de Passe</CardTitle>
                        </div>
                        <CardDescription>
                          Assurez-vous que votre compte utilise un mot de passe fort et unique.
                        </CardDescription>
                      </CardHeader>
                      <CardContent>
                        <form onSubmit={handlePasswordChange} className="space-y-4">
                          {passwordError && (
                            <div className="p-4 text-sm text-red-600 bg-red-50 dark:bg-red-900/10 border border-red-100 dark:border-red-900/20 rounded-xl">
                              {passwordError}
                            </div>
                          )}
                          {passwordSuccess && (
                            <div className="p-4 text-sm text-green-600 bg-green-50 dark:bg-green-900/10 border border-green-100 dark:border-green-900/20 rounded-xl">
                              {passwordSuccess}
                            </div>
                          )}
                          <div className="grid gap-4 sm:grid-cols-2">
                            <div className="space-y-2 sm:col-span-2">
                              <Label htmlFor="currentPassword">Mot de Passe Actuel</Label>
                              <Input
                                id="currentPassword"
                                type="password"
                                value={currentPassword}
                                onChange={(e) => setCurrentPassword(e.target.value)}
                                placeholder="••••••••"
                                required
                                disabled={isLoadingPassword}
                                className="bg-white dark:bg-black/20"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor="newPassword">Nouveau Mot de Passe</Label>
                              <Input
                                id="newPassword"
                                type="password"
                                value={newPassword}
                                onChange={(e) => setNewPassword(e.target.value)}
                                placeholder="8+ caractères"
                                required
                                disabled={isLoadingPassword}
                                minLength={8}
                                className="bg-white dark:bg-black/20"
                              />
                            </div>
                            <div className="space-y-2">
                              <Label htmlFor="confirmPassword">Confirmez le Mot de Passe</Label>
                              <Input
                                id="confirmPassword"
                                type="password"
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                placeholder="Confirmez le nouveau mot de passe"
                                required
                                disabled={isLoadingPassword}
                                minLength={8}
                                className="bg-white dark:bg-black/20"
                              />
                            </div>
                          </div>
                          <div className="pt-2">
                            <Button
                              type="submit"
                              disabled={isLoadingPassword}
                              className="bg-gradient-to-r from-red-500 to-red-600 hover:from-red-600 hover:to-red-700 text-white border-0"
                            >
                              {isLoadingPassword ? 'Changement...' : 'Changer le Mot de Passe'}
                            </Button>
                          </div>
                        </form>
                      </CardContent>
                    </Card>
                  </div>
                )}

                {activeTab === 'general' && (
                  <Card className="border-none shadow-xl bg-white/50 dark:bg-card/50 backdrop-blur-sm">
                    <CardHeader>
                      <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-blue-100 dark:bg-blue-900/30 text-blue-600 rounded-lg">
                          <User className="w-6 h-6" />
                        </div>
                        <CardTitle>Informations du Profil</CardTitle>
                      </div>
                      <CardDescription>Gérez les détails de votre profil public.</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                      <div className="flex items-center gap-4">
                        <div className="w-20 h-20 rounded-full bg-gradient-to-tr from-primary to-accent flex items-center justify-center text-white text-3xl font-bold">
                          {user?.username?.charAt(0).toUpperCase()}
                        </div>
                        <div>
                          <p className="font-bold text-lg">{user?.username}</p>
                          <p className="text-muted-foreground text-sm">Étudiant • Niveau 1</p>
                        </div>
                      </div>
                      <div className="p-4 bg-yellow-50 dark:bg-yellow-900/10 text-yellow-800 dark:text-yellow-200 rounded-xl text-sm">
                        👋 La personnalisation du profil arrive bientôt ! Vous pourrez changer votre avatar et votre nom d'affichage ici.
                      </div>
                    </CardContent>
                  </Card>
                )}

                {activeTab === 'appearance' && (
                  <Card className="border-none shadow-xl bg-white/50 dark:bg-card/50 backdrop-blur-sm">
                    <CardHeader>
                      <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-purple-100 dark:bg-purple-900/30 text-purple-600 rounded-lg">
                          <Palette className="w-6 h-6" />
                        </div>
                        <CardTitle>Thème & Apparence</CardTitle>
                      </div>
                      <CardDescription>Faites en sorte que l'application ressemble à ce que vous aimez !</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="flex flex-col gap-4">
                        <div className="flex items-center justify-between p-4 bg-white dark:bg-black/20 rounded-xl border border-border">
                          <span className="font-medium">Mode Sombre</span>
                          <ThemeToggle />
                        </div>
                        <div className="p-4 bg-purple-50 dark:bg-purple-900/10 text-purple-800 dark:text-purple-200 rounded-xl text-sm">
                          ✨ Plus de thèmes visuels (Espace, Jungle, Océan) arrivent dans la prochaine mise à jour !
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                )}

                {activeTab === 'notifications' && (
                  <Card className="border-none shadow-xl bg-white/50 dark:bg-card/50 backdrop-blur-sm">
                    <CardHeader>
                      <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-600 rounded-lg">
                          <Bell className="w-6 h-6" />
                        </div>
                        <CardTitle>Préférences de Notification</CardTitle>
                      </div>
                      <CardDescription>Contrôlez ce dont nous vous informons.</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="p-4 bg-gray-100 dark:bg-white/5 text-muted-foreground rounded-xl text-sm text-center">
                        Aucun paramètre de notification disponible pour le moment.
                      </div>
                    </CardContent>
                  </Card>
                )}

              </motion.div>
            </AnimatePresence>
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
