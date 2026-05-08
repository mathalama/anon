'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';
import { User, MessageSquare, Mic, Search as SearchIcon, X } from 'lucide-react';
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
    myGender, selectedGender, selectedMode,
    autoSearchOnReturn, setAutoSearchOnReturn,
    setMyGender, setSelectedGender, setSelectedMode
  } = useChat();
  const [isInitReady, setIsInitReady] = useState(false);
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

  const canStart = myGender !== '' && selectedGender !== '' && selectedMode !== '';

  useEffect(() => {
    resetSession();
  }, []);  // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    const init = async () => {
      let token = localStorage.getItem('access_token');
      if (token) {
        try {
          await api.getMe();
        } catch {
          localStorage.removeItem('access_token');
          token = null;
        }
      }
      if (!token) {
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
          const { access_token } = await api.createAnonymous(deviceId!);
          localStorage.setItem('access_token', access_token);
        } catch (e) {
          console.error('Init failed', e);
        }
      }
    };
    init().finally(() => setIsInitReady(true));
  }, []);

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

  if (status === 'searching') {
    return (
      <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-[#050505]">
        <div className="sleek-card max-w-sm w-full text-center space-y-10 py-12">
          <div className="relative mx-auto w-24 h-24">
            <div className="absolute inset-0 border-2 border-white/5 rounded-full" />
            <div className="absolute inset-0 border-2 border-blue-500 rounded-full border-t-transparent animate-spin" />
            <SearchIcon className="absolute inset-0 m-auto text-blue-500 w-8 h-8" />
          </div>
          <div className="space-y-3">
            <h2 className="text-2xl font-semibold tracking-tight">Поиск собеседника</h2>
            <p className="text-sm text-zinc-500 uppercase tracking-widest font-medium">
              {selectedMode} • {selectedGender}
            </p>
          </div>
          <button
            onClick={cancelSearch}
            className="sleek-button-secondary w-full"
          >
            Отмена
          </button>
        </div>
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
