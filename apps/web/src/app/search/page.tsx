'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';
import { User, MessageSquare, Mic, Search as SearchIcon } from 'lucide-react';
import { motion } from 'framer-motion';
import { Turnstile } from '@marsidev/react-turnstile';
import { clsx } from 'clsx';
import FingerprintJS from '@fingerprintjs/fingerprintjs';
import { api } from '@/lib/api';
import { useChatStore } from '@/store/chatStore';

type MyGender = 'male' | 'female';
type LookingFor = 'any' | 'male' | 'female';
type ChatMode = 'text' | 'voice';
type GenderOption = { id: MyGender; label: string };
type LookingForOption = { id: LookingFor; label: string };
type ModeOption = { id: ChatMode; label: string; icon: typeof MessageSquare };

export default function SearchPage() {
  const router = useRouter();
  const resetSession = useChatStore(s => s.resetSession);
  const {
    status, startSearch, cancelSearch,
    myGender, selectedGender, selectedMode, selectedTopic,
    autoSearchOnReturn, setAutoSearchOnReturn,
    setMyGender, setSelectedGender, setSelectedMode, setSelectedTopic
  } = useChat();
  const [isInitReady, setIsInitReady] = useState(false);
  const [needsCaptcha, setNeedsCaptcha] = useState(false);
  const [turnstileToken, setTurnstileToken] = useState('');

  const myGenderOptions: GenderOption[] = [
    { id: 'male', label: 'Парень' },
    { id: 'female', label: 'Девушка' },
  ];
  const lookingForOptions: LookingForOption[] = [
    { id: 'any', label: 'Любой' },
    { id: 'male', label: 'Парень' },
    { id: 'female', label: 'Девушка' },
  ];
  const modeOptions: ModeOption[] = [
    { id: 'text', label: 'Текст', icon: MessageSquare },
    { id: 'voice', label: 'Голос', icon: Mic },
  ];
  const topicOptions = [
    { id: 'global', label: 'Общая' },
    { id: 'anime', label: 'Аниме' },
    { id: 'roleplay', label: 'Ролевые' },
    { id: '18plus', label: '18+' },
  ];

  const canStart = myGender !== '' && selectedGender !== '' && selectedMode !== '' && selectedTopic !== '';

  useEffect(() => {
    resetSession();
  }, []);  // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    const init = async () => {
      let token = localStorage.getItem('access_token');
      if (token) {
        try {
          await api.getMe();
          setIsInitReady(true);
          return;
        } catch {
          localStorage.removeItem('access_token');
          token = null;
        }
      }
      if (!token) {
        setNeedsCaptcha(true);
      }
    };
    init();
  }, []);

  useEffect(() => {
    if (needsCaptcha && turnstileToken) {
      const doLogin = async () => {
        try {
          let deviceId = localStorage.getItem('device_id');
          if (!deviceId) {
            try {
              const fp = await FingerprintJS.load();
              const result = await fp.get();
              deviceId = result.visitorId;
            } catch {
              deviceId = Math.random().toString(36).substring(2) + Date.now().toString(36);
            }
            localStorage.setItem('device_id', deviceId!);
          }
          const { access_token } = await api.createAnonymous(deviceId!, turnstileToken);
          localStorage.setItem('access_token', access_token);
          setNeedsCaptcha(false);
          setIsInitReady(true);
        } catch (e) {
          console.error('Login failed', e);
        }
      };
      doLogin();
    }
  }, [needsCaptcha, turnstileToken]);

  useEffect(() => {
    if (status === 'chatting') {
      router.push('/chat');
    }
  }, [status, router]);

  const handleStart = () => {
    if (!canStart) return;
    setAutoSearchOnReturn(false);
    startSearch();
  };

  useEffect(() => {
    if (!isInitReady || !autoSearchOnReturn || !canStart || status !== 'idle') return;
    setAutoSearchOnReturn(false);
    startSearch();
  }, [isInitReady, autoSearchOnReturn, canStart, status, setAutoSearchOnReturn, startSearch]);

  if (needsCaptcha) {
    return (
      <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-[#050505]">
        <div className="sleek-card max-w-sm w-full text-center space-y-6 py-12">
          <h2 className="text-xl font-semibold tracking-tight">Защита от ботов</h2>
          <p className="text-zinc-500 text-sm">Пожалуйста, подтвердите, что вы человек.</p>
          <div className="flex justify-center pt-4">
            <Turnstile 
              siteKey={process.env.NEXT_PUBLIC_TURNSTILE_SITE_KEY || "1x00000000000000000000AA"} 
              onSuccess={setTurnstileToken} 
              options={{ theme: 'dark' }}
            />
          </div>
        </div>
      </main>
    );
  }

  if (status === 'searching') {
    return (
      <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-[#050505]">
        <motion.div 
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="sleek-card max-w-sm w-full text-center space-y-10 py-12"
        >
          <div className="relative mx-auto w-32 h-32 flex items-center justify-center">
            <motion.div 
              animate={{ scale: [1, 1.5, 1], opacity: [0.3, 0, 0.3] }}
              transition={{ repeat: Infinity, duration: 2, ease: "easeInOut" }}
              className="absolute inset-0 rounded-full bg-blue-500/30" 
            />
            <motion.div 
              animate={{ scale: [1, 1.2, 1], opacity: [0.5, 0, 0.5] }}
              transition={{ repeat: Infinity, duration: 2, delay: 0.5, ease: "easeInOut" }}
              className="absolute inset-0 rounded-full bg-blue-500/40" 
            />
            <div className="relative z-10 w-16 h-16 bg-[#050505] rounded-full flex items-center justify-center border border-white/10 shadow-xl shadow-blue-500/20">
               <SearchIcon className="text-blue-500 w-6 h-6 animate-pulse" />
            </div>
          </div>
          <div className="space-y-3">
            <h2 className="text-2xl font-semibold tracking-tight">Поиск собеседника</h2>
            <p className="text-sm text-zinc-500 uppercase tracking-widest font-medium">
              {selectedMode} • {selectedGender} • {topicOptions.find(t => t.id === selectedTopic)?.label}
            </p>
          </div>
          <button
            onClick={cancelSearch}
            className="sleek-button-secondary w-full"
          >
            Отмена
          </button>
        </motion.div>
      </main>
    );
  }

  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-[#050505]">
      <div className="max-w-md w-full space-y-12">
        <div className="text-center">
          <h1 className="text-3xl font-bold tracking-tight">Настройки поиска</h1>
          <p className="text-zinc-500 mt-2">Выберите параметры для начала чата</p>
        </div>

        <div className="space-y-8">
          {/* Room Topic */}
          <div className="space-y-4">
            <label className="text-xs font-semibold uppercase tracking-widest text-zinc-600">Комната —</label>
            <div className="grid grid-cols-2 gap-3">
              {topicOptions.map((item) => (
                <button
                  key={item.id}
                  onClick={() => setSelectedTopic(item.id)}
                  className={clsx(
                    "py-3 rounded-xl border transition-all font-medium text-sm",
                    selectedTopic === item.id
                      ? item.id === '18plus' 
                        ? "bg-red-600 border-red-500 text-white" 
                        : "bg-blue-600 border-blue-500 text-white"
                      : "bg-[#0f0f12] border-white/5 text-zinc-500 hover:border-white/10"
                  )}
                >
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          {/* I am */}
          <div className="space-y-4">
            <label className="text-xs font-semibold uppercase tracking-widest text-zinc-600">Я —</label>
            <div className="grid grid-cols-2 gap-3">
              {myGenderOptions.map((item) => (
                <button
                  key={item.id}
                  onClick={() => setMyGender(item.id)}
                  className={clsx(
                    "flex items-center justify-center space-x-3 py-4 rounded-xl border transition-all font-medium",
                    myGender === item.id
                      ? "bg-blue-600 border-blue-500 text-white"
                      : "bg-[#0f0f12] border-white/5 text-zinc-500 hover:border-white/10"
                  )}
                >
                  <User className="w-5 h-5" />
                  <span>{item.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Looking for */}
          <div className="space-y-4">
            <label className="text-xs font-semibold uppercase tracking-widest text-zinc-600">Ищу —</label>
            <div className="grid grid-cols-3 gap-3">
              {lookingForOptions.map((item) => (
                <button
                  key={item.id}
                  onClick={() => setSelectedGender(item.id)}
                  className={clsx(
                    "py-3 rounded-xl border transition-all font-medium text-sm",
                    selectedGender === item.id
                      ? "bg-blue-600 border-blue-500 text-white"
                      : "bg-[#0f0f12] border-white/5 text-zinc-500 hover:border-white/10"
                  )}
                >
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          {/* Mode */}
          <div className="space-y-4">
            <label className="text-xs font-semibold uppercase tracking-widest text-zinc-600">Формат —</label>
            <div className="grid grid-cols-2 gap-3">
              {modeOptions.map((item) => (
                <button
                  key={item.id}
                  onClick={() => setSelectedMode(item.id)}
                  className={clsx(
                    "flex items-center justify-center space-x-3 py-4 rounded-xl border transition-all font-medium",
                    selectedMode === item.id
                      ? "bg-blue-600 border-blue-500 text-white"
                      : "bg-[#0f0f12] border-white/5 text-zinc-500 hover:border-white/10"
                  )}
                >
                  <item.icon className="w-5 h-5" />
                  <span>{item.label}</span>
                </button>
              ))}
            </div>
          </div>

          <div className="pt-6">
            <button
              onClick={handleStart}
              disabled={!canStart}
              className={clsx(
                "w-full py-5 rounded-2xl font-semibold text-lg transition-all",
                canStart
                  ? "bg-blue-600 text-white hover:brightness-110"
                  : "bg-zinc-900 text-zinc-700 cursor-not-allowed border border-white/5"
              )}
            >
              {canStart ? 'Найти собеседника' : 'Заполни все поля'}
            </button>
          </div>
        </div>
      </div>
    </main>
  );
}
