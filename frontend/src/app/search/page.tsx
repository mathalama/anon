'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';
import { User, MessageSquare, Mic, Search as SearchIcon, X } from 'lucide-react';
import { clsx } from 'clsx';
import FingerprintJS from '@fingerprintjs/fingerprintjs';
import { api } from '@/lib/api';

export default function SearchPage() {
  const router = useRouter();
  const { status, startSearch, cancelSearch, myGender, selectedGender, selectedMode, setMyGender, setSelectedGender, setSelectedMode } = useChat();

  useEffect(() => {
    const init = async () => {
      let token = localStorage.getItem('access_token');
      
      // Check if token works
      if (token) {
        try {
          await api.getMe();
        } catch (e) {
          console.warn('Token expired or invalid, clearing...');
          localStorage.removeItem('access_token');
          token = null;
        }
      }

      if (!token) {
        console.log('Creating fresh anonymous session...');
        try {
          let deviceId = localStorage.getItem('device_id');
          if (!deviceId) {
            // Use fingerprint or random UUID
            try {
              const fp = await FingerprintJS.load();
              const result = await fp.get();
              deviceId = result.visitorId;
            } catch (e) {
              deviceId = Math.random().toString(36).substring(2) + Date.now().toString(36);
            }
            localStorage.setItem('device_id', deviceId!);
          }
          
          const { access_token } = await api.createAnonymous(deviceId!);
          localStorage.setItem('access_token', access_token);
          console.log('New session initialized');
        } catch (e) {
          console.error('Initialization failed', e);
        }
      }
    };
    init();
  }, []);

  useEffect(() => {
    if (status === 'matched') {
      router.push('/chat');
    }
  }, [status, router]);

  const handleStart = () => {
    startSearch();
  };

  if (status === 'searching') {
    return (
      <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-white dark:bg-[#1c1c1c]">
        <div className="text-center space-y-6">
          <div className="relative">
            <div className="w-20 h-20 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto"></div>
            <SearchIcon className="absolute inset-0 m-auto text-blue-500 w-8 h-8" />
          </div>
          <div className="space-y-2">
            <h2 className="text-2xl font-bold">Searching for someone...</h2>
            <p className="text-gray-500">Looking for a match based on your filters</p>
          </div>
          <button 
            onClick={cancelSearch}
            className="px-8 py-2 border border-gray-300 dark:border-gray-700 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            Cancel
          </button>
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-white dark:bg-[#1c1c1c]">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <h1 className="text-2xl font-bold">Search Settings</h1>
        </div>

        <div className="space-y-6">
          {/* I am... */}
          <div className="space-y-3">
            <label className="text-sm font-medium text-gray-500 uppercase">I am</label>
            <div className="grid grid-cols-2 gap-4">
              {[
                { id: 'male', label: 'Guy', icon: User },
                { id: 'female', label: 'Girl', icon: User },
              ].map((item) => (
                <button
                  key={item.id}
                  onClick={() => setMyGender(item.id as any)}
                  className={clsx(
                    "flex items-center justify-center space-x-2 p-4 rounded-xl border transition-all",
                    myGender === item.id 
                      ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-blue-600" 
                      : "border-gray-200 dark:border-gray-800 text-gray-500"
                  )}
                >
                  <item.icon className="w-5 h-5" />
                  <span className="font-medium">{item.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Looking for... */}
          <div className="space-y-3">
            <label className="text-sm font-medium text-gray-500 uppercase">Looking for</label>
            <div className="grid grid-cols-3 gap-2">
              {[
                { id: 'any', label: 'Any' },
                { id: 'male', label: 'Guy' },
                { id: 'female', label: 'Girl' },
              ].map((item) => (
                <button
                  key={item.id}
                  onClick={() => setSelectedGender(item.id as any)}
                  className={clsx(
                    "p-3 rounded-xl border text-sm font-medium transition-all",
                    selectedGender === item.id 
                      ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-blue-600" 
                      : "border-gray-200 dark:border-gray-800 text-gray-500"
                  )}
                >
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          {/* Mode */}
          <div className="space-y-3">
            <label className="text-sm font-medium text-gray-500 uppercase">Chat Mode</label>
            <div className="grid grid-cols-2 gap-4">
              {[
                { id: 'text', label: 'Text', icon: MessageSquare },
                { id: 'voice', label: 'Voice', icon: Mic },
              ].map((item) => (
                <button
                  key={item.id}
                  onClick={() => setSelectedMode(item.id as any)}
                  className={clsx(
                    "flex items-center justify-center space-x-2 p-4 rounded-xl border transition-all",
                    selectedMode === item.id 
                      ? "border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-blue-600" 
                      : "border-gray-200 dark:border-gray-800 text-gray-500"
                  )}
                >
                  <item.icon className="w-5 h-5" />
                  <span className="font-medium">{item.label}</span>
                </button>
              ))}
            </div>
          </div>

          <button 
            onClick={handleStart}
            className="w-full py-4 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-bold transition-colors"
          >
            Find a Partner
          </button>
        </div>
      </div>
    </main>
  );
}
