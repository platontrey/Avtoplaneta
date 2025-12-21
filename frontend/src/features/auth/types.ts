/*
* Copyright (c) 2025 Avtoplaneta. All rights reserved.
*/

export interface User {
  id: number;
  email: string;
  name: string;
  provider: string;
  role: string;
  initials?: string;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
}
