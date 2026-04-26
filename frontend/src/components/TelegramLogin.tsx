'use client';

import { useEffect, useRef } from 'react';
import { api } from '@/lib/api';
import { useRouter } from 'next/navigation';

export default function TelegramLogin() {
  const router = useRouter();
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    console.log('TelegramLogin: Initializing...');

    // Define the callback function globally
    (window as any).onTelegramAuth = async (user: any) => {
      console.log('Telegram Auth Received:', user);
      try {
        const data = await api.loginTelegram(user);
        localStorage.setItem('access_token', data.access_token);
        localStorage.setItem('refresh_token', data.refresh_token);
        router.push('/search');
      } catch (err) {
        console.error('Telegram Auth API Error:', err);
        alert('Authentication failed on server');
      }
    };

    // Create script element
    const script = document.createElement('script');
    script.src = 'https://telegram.org/js/telegram-widget.js?22';
    script.async = true;
    script.setAttribute('data-telegram-login', 'nektomekz_bot');
    script.setAttribute('data-size', 'large');
    script.setAttribute('data-onauth', 'onTelegramAuth(user)');
    script.setAttribute('data-request-access', 'write');

    // Clear and append
    if (containerRef.current) {
      containerRef.current.innerHTML = '';
      containerRef.current.appendChild(script);
      console.log('TelegramLogin: Script appended to DOM');
    }

    return () => {
      console.log('TelegramLogin: Cleaning up...');
      if (containerRef.current) {
        containerRef.current.innerHTML = '';
      }
      // Keep the callback global to avoid race conditions if widget stays in cache
    };
  }, [router]);

  return (
    <div className="min-h-[40px] flex items-center justify-center">
      <div ref={containerRef} id="telegram-login-container" />
    </div>
  );
}
