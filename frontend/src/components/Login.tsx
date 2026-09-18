/*
 * Copyright (c) 2025 Avtoplaneta. All rights reserved.
 */

import { useState } from 'react';
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Eye, EyeOff } from 'lucide-react';
import ThemeToggle from './ThemeToggle';

import { useAuth } from '../features/auth/hooks/useAuth';

const loginSchema = z.object({
  email: z.string().min(1, { message: "Введите электронную почту или имя пользователя" }),
  password: z.string().min(1, { message: "Введите пароль" }),
});

type LoginFormData = z.infer<typeof loginSchema>;

export default function Login() {
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const { login, isLoading } = useAuth();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
  });

  const handleGoogleLogin = () => {
    window.location.assign('/auth/google');
  };

  const onSubmit = async (data: LoginFormData) => {
    setError('');

    console.log('Login attempt:', { email: data.email, password: '[HIDDEN]' });

    try {
      await login(data);
      // Redirect to inventory after successful login
      window.location.href = '/';
    } catch (err) {
      console.error('Login error:', err);
      setError(err instanceof Error ? err.message : 'Не удалось войти');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background py-12 px-4 sm:px-6 lg:px-8 relative">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>

      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-foreground">
            Вход в личный кабинет
          </h2>
          <p className="mt-2 text-center text-sm text-muted-foreground">
            Система управления складом автозапчастей
          </p>
        </div>

        <div className="space-y-6">
          {/* Login Form */}
          <Card>
            <CardHeader>
              <CardTitle className="text-center">
                Вход в систему
              </CardTitle>
              <CardDescription className="text-center">
                Введите электронную почту или имя пользователя и пароль
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
                <div>
                  <Label htmlFor="email">Электронная почта или имя пользователя</Label>
                  <Input
                    id="email"
                    {...register("email")}
                    type="text"
                    autoComplete="email"
                    placeholder="Введите электронную почту или имя пользователя"
                  />
                  {errors.email && <p className="text-red-500 text-sm">{errors.email.message}</p>}
                </div>
                <div>
                  <Label htmlFor="password">
                    Пароль
                  </Label>
                  <div className="relative">
                    <Input
                      id="password"
                      {...register("password")}
                      type={showPassword ? "text" : "password"}
                      autoComplete="current-password"
                      placeholder="Введите пароль"
                      className="pr-10"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                      onClick={() => setShowPassword(!showPassword)}
                      aria-label={showPassword ? "Скрыть пароль" : "Показать пароль"}
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </Button>
                  </div>
                  {errors.password && <p className="text-red-500 text-sm">{errors.password.message}</p>}
                </div>

                {error && (
                  <div className="text-red-600 text-sm text-center">
                    {error}
                  </div>
                )}

                <Button
                  type="submit"
                  className="w-full"
                  disabled={isLoading}
                >
                  {isLoading ? 'Выполняется вход…' : 'Войти'}
                </Button>
              </form>
            </CardContent>
          </Card>

          {/* Divider */}
          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <div className="w-full border-t border-border" />
            </div>
            <div className="relative flex justify-center text-sm">
              <span className="px-2 bg-background text-muted-foreground">или</span>
            </div>
          </div>

          {/* Google Login */}
          <Card>
            <CardHeader>
              <CardTitle className="text-center">Вход через Google</CardTitle>
              <CardDescription className="text-center">
                Используйте свою учётную запись Google
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Button
                onClick={handleGoogleLogin}
                className="w-full"
                variant="outline"
              >
                <svg className="w-5 h-5 mr-2" viewBox="0 0 24 24">
                  <path
                    fill="currentColor"
                    d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
                  />
                  <path
                    fill="currentColor"
                    d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
                  />
                  <path
                    fill="currentColor"
                    d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
                  />
                  <path
                    fill="currentColor"
                    d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
                  />
                </svg>
                Продолжить через Google
              </Button>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
