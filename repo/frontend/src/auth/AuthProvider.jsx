import React, { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { login as apiLogin, logout as apiLogout, revokeToken, me } from '../lib/api';

const IDLE_LIMIT_MS = 30 * 60 * 1000;
const WARNING_MS = 2 * 60 * 1000;

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(() => {
    const raw = localStorage.getItem('user');
    return raw ? JSON.parse(raw) : null;
  });
  const [loading, setLoading] = useState(Boolean(localStorage.getItem('token')));
  const [warningOpen, setWarningOpen] = useState(false);
  const [secondsLeft, setSecondsLeft] = useState(0);
  const lastActivity = useRef(Date.now());

  useEffect(() => {
    if (!localStorage.getItem('token')) {
      setLoading(false);
      return;
    }
    me()
      .then((profile) => {
        setUser(profile);
        localStorage.setItem('user', JSON.stringify(profile));
      })
      .catch(() => {
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        setUser(null);
      })
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    const markActivity = () => {
      lastActivity.current = Date.now();
      setWarningOpen(false);
    };
    window.addEventListener('mousemove', markActivity);
    window.addEventListener('keydown', markActivity);
    window.addEventListener('click', markActivity);
    window.addEventListener('scroll', markActivity);
    return () => {
      window.removeEventListener('mousemove', markActivity);
      window.removeEventListener('keydown', markActivity);
      window.removeEventListener('click', markActivity);
      window.removeEventListener('scroll', markActivity);
    };
  }, []);

  useEffect(() => {
    if (!user) {
      return undefined;
    }
    const timer = setInterval(async () => {
      const idleMs = Date.now() - lastActivity.current;
      if (idleMs >= IDLE_LIMIT_MS) {
        // Optimistic logout: clear client state immediately, then
        // attempt to revoke the token in background so UI never blocks
        const token = localStorage.getItem('token');
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        setUser(null);
        setWarningOpen(false);
        setSecondsLeft(0);
        revokeToken(token);
        return;
      }
      const msLeft = IDLE_LIMIT_MS - idleMs;
      if (msLeft <= WARNING_MS) {
        setWarningOpen(true);
        setSecondsLeft(Math.max(0, Math.floor(msLeft / 1000)));
      }
    }, 1000);
    return () => clearInterval(timer);
  }, [user]);

    const value = useMemo(() => ({
    user,
    loading,
    warningOpen,
    secondsLeft,
    async login(payload) {
      const data = await apiLogin(payload);
      setUser(data.user);
      lastActivity.current = Date.now();
      return data;
    },
    async logout() {
      // Optimistic logout: immediately clear client state so user is signed out
      const token = localStorage.getItem('token');
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      setUser(null);
      setWarningOpen(false);
      // Revoke token in background; do not block the UI
      revokeToken(token);
    },
    async refreshMe() {
      const profile = await me();
      setUser(profile);
      localStorage.setItem('user', JSON.stringify(profile));
      return profile;
    },
    staySignedIn() {
      lastActivity.current = Date.now();
      setWarningOpen(false);
    },
  }), [user, loading, warningOpen, secondsLeft]);

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  return useContext(AuthContext);
}

