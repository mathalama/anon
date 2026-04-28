'use client';

import Link from 'next/link';
import { MessageSquare, Mic, Shield } from 'lucide-react';

export default function Home() {
  return (
    <main className="min-h-screen flex flex-col items-center justify-center p-6 bg-white dark:bg-[#1c1c1c]">
      <div className="max-w-md w-full text-center space-y-8">
        <div className="space-y-4">
          <h1 className="text-4xl font-bold tracking-tight text-gray-900 dark:text-white">
            NektoKZ
          </h1>
          <p className="text-lg text-gray-600 dark:text-gray-400">
            Anonymous chat for everyone. Simple, fast and safe.
          </p>
        </div>

        <div className="grid grid-cols-1 gap-4 text-left">
          <div className="flex items-center space-x-4 p-4 border border-gray-200 dark:border-gray-800 rounded-xl">
            <MessageSquare className="text-blue-500" />
            <div>
              <h3 className="font-semibold">Text Chat</h3>
              <p className="text-sm text-gray-500">Traditional anonymous messaging</p>
            </div>
          </div>
          <div className="flex items-center space-x-4 p-4 border border-gray-200 dark:border-gray-800 rounded-xl">
            <Mic className="text-green-500" />
            <div>
              <h3 className="font-semibold">Voice Chat</h3>
              <p className="text-sm text-gray-500">Real-time voice calls</p>
            </div>
          </div>
        </div>

        <Link 
          href="/search" 
          className="block w-full py-4 bg-blue-500 hover:bg-blue-600 text-white rounded-xl font-bold transition-colors text-center"
        >
          Start Chatting
        </Link>
        
        <p className="text-xs text-gray-400">
          By clicking start, you agree to our terms of service.
        </p>
      </div>
    </main>
  );
}
