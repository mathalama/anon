'use client';

import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';
import { useVoiceCall } from '@/hooks/useVoiceCall';
import { Send, User, ChevronLeft, Flag, MoreVertical } from 'lucide-react';
import { VoiceCallUI } from '@/components/VoiceCallUI';
import { clsx } from 'clsx';

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
    sendMessage, 
    sendTyping, 
    next 
  } = useChat();

  const { callState, isMuted, toggleMute, endCall, remoteAudioRef } = useVoiceCall();

  useEffect(() => {
    if (status === 'idle') {
      router.push('/');
    }
  }, [status, router]);

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
    next();
    router.push('/search');
  };

  const handleEndCall = () => {
    endCall();
    next();
    router.push('/');
  };

  return (
    <main className="flex flex-col h-screen bg-[#f0f2f5] dark:bg-[#121212]">
      {/* Simple Header */}
      <header className="flex items-center justify-between px-4 py-3 bg-white dark:bg-[#1c1c1c] border-b border-gray-200 dark:border-gray-800 shadow-sm">
        <div className="flex items-center space-x-3">
          <button onClick={handleEndCall} className="p-1 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full">
            <ChevronLeft className="w-6 h-6" />
          </button>
          <div className="w-10 h-10 rounded-full bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
            <User className="text-gray-500 w-6 h-6" />
          </div>
          <div>
            <h2 className="font-semibold text-sm">
              Stranger ({partnerGender || '...'})
            </h2>
            <div className="flex items-center space-x-1">
              <span className="w-2 h-2 bg-green-500 rounded-full"></span>
              <span className="text-[10px] text-gray-500 uppercase font-medium">Online</span>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <button 
            onClick={handleNext}
            className="px-4 py-1.5 bg-blue-500 hover:bg-blue-600 text-white text-sm font-bold rounded-lg transition-colors"
          >
            Next
          </button>
          <button className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-full text-gray-400">
            <Flag className="w-5 h-5" />
          </button>
        </div>
      </header>

      {/* Chat Content */}
      <div className="flex-1 relative overflow-hidden flex flex-col">
        {mode === 'voice' && (
          <div className="absolute inset-0 z-10 bg-white dark:bg-[#1c1c1c]">
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

        {/* Message Area */}
        <div 
          ref={scrollRef}
          className="flex-1 overflow-y-auto p-4 space-y-4 scroll-smooth"
        >
          {messages.map((msg, i) => (
            <div 
              key={i}
              className={clsx(
                "flex w-full",
                msg.sender === 'me' ? "justify-end" : "justify-start"
              )}
            >
              <div 
                className={clsx(
                  "max-w-[80%] px-4 py-2.5 shadow-sm",
                  msg.sender === 'me' 
                    ? "bg-blue-500 text-white rounded-2xl rounded-tr-none" 
                    : "bg-white dark:bg-[#1c1c1c] text-gray-900 dark:text-gray-100 rounded-2xl rounded-tl-none border border-gray-100 dark:border-gray-800"
                )}
              >
                <p className="text-sm leading-relaxed">{msg.content}</p>                <span className="text-[10px] opacity-70 block mt-1 text-right">
                  {new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
            </div>
          ))}

          {isPartnerTyping && (
            <div className="flex justify-start">
              <div className="bg-white dark:bg-[#1c1c1c] px-4 py-3 rounded-2xl border border-gray-100 dark:border-gray-800 shadow-sm">
                <div className="flex space-x-1">
                  <div className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce"></div>
                  <div className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce [animation-delay:0.2s]"></div>
                  <div className="w-1.5 h-1.5 bg-gray-400 rounded-full animate-bounce [animation-delay:0.4s]"></div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Input Area */}
        <form 
          onSubmit={handleSend}
          className="p-4 bg-white dark:bg-[#1c1c1c] border-t border-gray-200 dark:border-gray-800"
        >
          <div className="flex items-center space-x-2">
            <input 
              type="text"
              value={input}
              onChange={handleInputChange}
              placeholder="Type a message..."
              className="flex-1 bg-gray-100 dark:bg-[#2c2c2c] border-none rounded-xl px-4 py-3 text-sm focus:ring-2 focus:ring-blue-500 outline-none"
            />
            <button 
              type="submit"
              disabled={!input.trim()}
              className="p-3 bg-blue-500 text-white rounded-xl hover:bg-blue-600 disabled:opacity-50 disabled:hover:bg-blue-500 transition-colors shadow-sm"
            >
              <Send className="w-5 h-5" />
            </button>
          </div>
        </form>
      </div>
    </main>
  );
}
