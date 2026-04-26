'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';

export default function SearchPage() {
  const router = useRouter();
  const { status, startSearch, cancelSearch } = useChat();

  useEffect(() => {
    if (!localStorage.getItem('access_token')) {
      router.push('/');
      return;
    }

    if (status === 'idle') {
      startSearch();
    } else if (status === 'matched' || status === 'chatting') {
      router.push('/chat');
    }
  }, [status, startSearch, router]);

  const handleCancel = async () => {
    await cancelSearch();
    router.push('/');
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-6 text-center bg-slate-950">
      <div className="relative w-64 h-64 flex items-center justify-center">
        {/* Pulsing background circles */}
        <div className="absolute inset-0 rounded-full bg-indigo-500/20 animate-ping duration-1000"></div>
        <div className="absolute inset-4 rounded-full bg-indigo-500/30 animate-pulse duration-700"></div>
        
        {/* Radar lines */}
        <div className="absolute inset-0 border-2 border-indigo-500/20 rounded-full"></div>
        <div className="absolute inset-10 border border-indigo-500/10 rounded-full"></div>
        
        <div className="relative z-10 p-8 rounded-full bg-indigo-600 shadow-2xl shadow-indigo-500/50">
          <svg className="w-16 h-16 text-white animate-bounce" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>

      <div className="mt-12 space-y-4">
        <h2 className="text-3xl font-bold">Searching...</h2>
        <p className="text-slate-400 max-w-xs mx-auto">
          We're finding the perfect person for you to talk to. This usually takes just a few seconds.
        </p>
      </div>

      <button
        onClick={handleCancel}
        className="mt-12 px-6 py-2 rounded-xl border border-slate-800 text-slate-400 hover:text-slate-100 hover:bg-slate-900 transition-colors"
      >
        Cancel Search
      </button>
    </main>
  );
}
