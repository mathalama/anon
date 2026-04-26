'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import FingerprintJS from '@fingerprintjs/fingerprintjs';
import TelegramLogin from '@/components/TelegramLogin';

export default function LandingPage() {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);

  const handleStart = async () => {
    console.log('handleStart: Clicked');
    setIsLoading(true);
    try {
      console.log('handleStart: Loading Fingerprint...');
      const fp = await FingerprintJS.load();
      const result = await fp.get();
      const deviceId = result.visitorId;
      console.log('handleStart: Device ID generated:', deviceId);
      
      localStorage.setItem('device_id', deviceId);

      console.log('handleStart: Calling createAnonymous API...');
      const data = await api.createAnonymous(deviceId);
      console.log('handleStart: API Success, tokens received');
      
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      
      console.log('handleStart: Redirecting to /search');
      router.push('/search');
    } catch (err) {
      console.error('handleStart ERROR:', err);
      alert('Failed to connect to server. Check console for details.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-6 text-center bg-gradient-to-b from-slate-950 to-indigo-950">
      <div className="max-w-md space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-1000">
        <div className="space-y-4">
          <div className="inline-block p-3 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 mb-4">
            <svg className="w-12 h-12 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
            </svg>
          </div>
          <h1 className="text-5xl font-extrabold tracking-tight sm:text-6xl">
            Nekto<span className="text-indigo-500">KZ</span>
          </h1>
          <p className="text-xl text-slate-400">
            Find someone to talk to, anonymously. Safe, fast, and free.
          </p>
        </div>

        <div className="pt-8 space-y-4">
          <button
            onClick={handleStart}
            disabled={isLoading}
            className="group relative w-full inline-flex items-center justify-center px-8 py-4 font-bold text-white transition-all duration-200 bg-indigo-600 rounded-2xl hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-600 shadow-lg shadow-indigo-500/25 active:scale-95 disabled:opacity-50 disabled:pointer-events-none"
          >
            {isLoading ? (
              <span className="flex items-center">
                <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                Connecting...
              </span>
            ) : (
              "Start Anonymously"
            )}
          </button>

          <div className="relative">
            <div className="absolute inset-0 flex items-center">
              <span className="w-full border-t border-slate-800"></span>
            </div>
            <div className="relative flex justify-center text-xs uppercase">
              <span className="bg-slate-950 px-2 text-slate-500">Or use stable account</span>
            </div>
          </div>

          <div className="flex justify-center">
             <TelegramLogin />
          </div>
          
          <p className="mt-4 text-sm text-slate-500">
            By starting, you agree to our terms and community guidelines.
          </p>
        </div>
      </div>
      
      <div className="absolute bottom-8 text-slate-600 text-sm">
        &copy; 2026 NektoKZ Team. Made with &hearts;
      </div>
    </main>
  );
}
