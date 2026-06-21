'use client';

import { useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';
import { useVoiceCall } from '@/hooks/useVoiceCall';
import { User, ChevronLeft, Flag } from 'lucide-react';
import dynamic from 'next/dynamic';
import { clsx } from 'clsx';
import { api } from '@/lib/api';
import { chatSocket } from '@/lib/socket';
import { MessageBubble } from '@/components/MessageBubble';
import { ChatInput } from '@/components/ChatInput';
import { playSound } from '@/lib/sounds';

const VoiceCallUI = dynamic(() => import('@/components/VoiceCallUI').then(mod => mod.VoiceCallUI), { ssr: false });


export default function ChatPage() {
  const router = useRouter();
  const scrollRef = useRef<HTMLDivElement>(null);

  const {
    status,
    mode,
    messages,
    isPartnerTyping,
    partnerGender,
    partnerUserId,
    roomId,

    setAutoSearchOnReturn,
    sendMessage,
    sendTyping,
    next,
  } = useChat();

  const { callState, isMuted, toggleMute, endCall, remoteAudioRef } = useVoiceCall();

  useEffect(() => {
    return () => {
      endCall();
      chatSocket.disconnect();
    };
  }, [endCall]);

  useEffect(() => {
    if (status === 'idle') {
      router.push('/');
    }
    if (status === 'ended') {
      setAutoSearchOnReturn(true);
      router.push('/search');
    }
  }, [status, setAutoSearchOnReturn, router]);

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [messages, isPartnerTyping]);

  const prevStatus = useRef(status);
  const prevMessagesLen = useRef(messages.length);

  useEffect(() => {
    if (prevStatus.current === 'searching' && status === 'chatting') {
      playSound('match');
    } else if (prevStatus.current === 'chatting' && status === 'ended') {
      playSound('disconnect');
    }
    prevStatus.current = status;
  }, [status]);

  useEffect(() => {
    if (messages.length > prevMessagesLen.current) {
      const lastMsg = messages[messages.length - 1];
      if (lastMsg && lastMsg.sender !== 'me') {
        playSound('message');
      }
    }
    prevMessagesLen.current = messages.length;
  }, [messages]);

  const handleSend = (content: string) => {
    sendMessage(content);
    sendTyping(false);
  };

  const handleTyping = (isTyping: boolean) => {
    sendTyping(isTyping);
  };

  const handleNext = () => {
    setAutoSearchOnReturn(true);
    next();
    router.push('/search');
  };

  const handleEndCall = () => {
    endCall();
    setAutoSearchOnReturn(true);
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
            <MessageBubble key={i} msg={msg} />
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

        <ChatInput onSendMessage={handleSend} onTyping={handleTyping} />
      </div>
    </main>
  );
}

