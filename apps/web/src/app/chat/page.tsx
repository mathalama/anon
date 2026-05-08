'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';
import { useVoiceCall } from '@/hooks/useVoiceCall';
import { Send, User, ChevronLeft, Flag, X } from 'lucide-react';
import { VoiceCallUI } from '@/components/VoiceCallUI';
import { clsx } from 'clsx';
import { api } from '@/lib/api';

export default function ChatPage() {
  const router = useRouter();
  const [input, setInput] = useState('');
  const scrollRef = useRef<HTMLDivElement>(null);

  const {
    status,
    mode,
    messages,
    isPartnerTyping,
    partnerGender,
    partnerUserId,
    roomId,
    endReason,
    setAutoSearchOnReturn,
    sendMessage,
    sendTyping,
    next,
  } = useChat();

  const { callState, isMuted, toggleMute, endCall, remoteAudioRef } = useVoiceCall();

  useEffect(() => {
    if (status === 'idle') {
      router.push('/');
    }
    if (status === 'ended') {
      if (endReason === 'disconnect') {
        setAutoSearchOnReturn(true);
      }
      router.push('/search');
    }
  }, [status, endReason, setAutoSearchOnReturn, router]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, isPartnerTyping]);

  const handleSend = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (input.trim()) {
      sendMessage(input.trim());
      setInput('');
      sendTyping(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value);
    sendTyping(e.target.value.length > 0);
  };

  const handleNext = () => {
    setAutoSearchOnReturn(false);
    next();
    router.push('/search');
  };

  const handleEndCall = () => {
    endCall();
    setAutoSearchOnReturn(false);
    next();
    router.push('/search');
  };

  const handleReport = async () => {
    if (!roomId) return;

    const reason = window.prompt('Report reason (short):', 'abuse')?.trim() || 'abuse';
    if (!partnerUserId) {
      alert('Cannot send report: backend did not return partner_user_id for this session.');
      return;
    }

    try {
      await api.reportUser(roomId, partnerUserId, reason);
      alert('Report submitted.');
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : 'Failed to submit report';
      alert(message);
    }
  };

  return (
    <main className="flex flex-col h-screen text-white bg-[#050505] overflow-hidden">
      <header className="flex items-center justify-between px-6 py-4 border-b border-white/5 bg-[#050505]/80 backdrop-blur-md z-20 relative">
        <div className="flex items-center space-x-4">
          <button
            onClick={handleEndCall}
            className="p-2.5 rounded-xl bg-zinc-900 border border-white/5 text-zinc-400 hover:text-white transition-colors"
          >
            <ChevronLeft className="w-6 h-6" />
          </button>
          <div className="relative">
            <div className="w-12 h-12 bg-blue-600/10 border border-blue-600/20 rounded-xl flex items-center justify-center">
              <User className="text-blue-500 w-7 h-7" />
            </div>
            <div className="absolute -bottom-1 -right-1 w-3.5 h-3.5 bg-emerald-500 border-2 border-[#050505] rounded-full"></div>
          </div>
          <div>
            <h2 className="font-semibold text-lg tracking-tight">
              Собеседник <span className="text-zinc-500 ml-1 text-sm">({partnerGender || '...'})</span>
            </h2>
            <div className="flex items-center space-x-1.5">
              <span
                className={clsx(
                  'text-[10px] uppercase font-bold tracking-widest',
                  status === 'chatting' ? 'text-emerald-500' : 'text-amber-500',
                )}
              >
                {status === 'chatting' ? 'В сети' : 'Поиск...'}
              </span>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          <button
            onClick={handleNext}
            className="sleek-button px-6 py-2.5 text-sm font-semibold"
          >
            Следующий
          </button>
          <button
            onClick={handleReport}
            className="p-2.5 rounded-xl bg-zinc-900 border border-white/5 text-zinc-500 hover:text-red-500 hover:bg-red-500/10 transition-all"
          >
            <Flag className="w-5 h-5" />
          </button>
        </div>
      </header>

      <div className="flex-1 relative overflow-hidden flex flex-col">
        {mode === 'voice' && (
          <div className="absolute inset-0 z-10 bg-[#050505] flex flex-col">
            <VoiceCallUI
              callState={callState}
              isMuted={isMuted}
              toggleMute={toggleMute}
              endCall={handleEndCall}
              remoteAudioRef={remoteAudioRef}
              partnerGender={partnerGender}
            />
          </div>
        )}

        <div ref={scrollRef} className="flex-1 overflow-y-auto p-6 space-y-6 scroll-smooth">
          {messages.map((msg, i) => (
            <div
              key={i}
              className={clsx(
                'flex w-full',
                msg.sender === 'me' ? 'justify-end' : 'justify-start',
              )}
            >
              <div
                className={clsx(
                  'chat-bubble',
                  msg.sender === 'me' ? 'chat-bubble-me' : 'chat-bubble-partner',
                )}
              >
                <p>{msg.content}</p>
                <span
                  className={clsx(
                    'text-[10px] font-medium uppercase tracking-widest mt-2 block opacity-40',
                    msg.sender === 'me' ? 'text-right' : 'text-left',
                  )}
                >
                  {new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
            </div>
          ))}

          {isPartnerTyping && (
            <div className="flex justify-start">
              <div className="chat-bubble chat-bubble-partner">
                <div className="flex space-x-1.5 py-1">
                  <div className="w-1.5 h-1.5 bg-zinc-500 rounded-full animate-bounce"></div>
                  <div className="w-1.5 h-1.5 bg-zinc-500 rounded-full animate-bounce [animation-delay:0.2s]"></div>
                  <div className="w-1.5 h-1.5 bg-zinc-500 rounded-full animate-bounce [animation-delay:0.4s]"></div>
                </div>
              </div>
            </div>
          )}
        </div>

        <form onSubmit={handleSend} className="p-6 border-t border-white/5 bg-[#050505] mt-auto">
          <div className="flex items-center gap-3 max-w-5xl mx-auto">
            <input
              type="text"
              value={input}
              onChange={handleInputChange}
              placeholder="Напишите сообщение..."
              className="sleek-input flex-1"
            />
            <button
              type="submit"
              disabled={!input.trim()}
              className="sleek-button p-4 rounded-xl"
            >
              <Send className="w-6 h-6" />
            </button>
          </div>
        </form>
      </div>
    </main>
  );
}

