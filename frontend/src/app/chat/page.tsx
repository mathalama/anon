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
    if (status === 'ended') {
        // показать уведомление и редиректнуть
        router.push('/search');
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
    router.push('/search');
  };

  return (
    <main className="flex flex-col h-screen text-foreground overflow-hidden">
      {/* Premium Glass Header */}
      <header className="flex items-center justify-between px-6 py-4 glass z-20 relative">
        <div className="flex items-center space-x-4">
          <button 
            onClick={handleEndCall} 
            className="p-2 hover:bg-foreground/5 rounded-xl transition-all active:scale-95"
          >
            <ChevronLeft className="w-6 h-6" />
          </button>
          <div className="relative">
            <div className="w-11 h-11 rounded-2xl bg-gradient-to-br from-primary/20 to-accent/20 flex items-center justify-center border border-primary/20">
              <User className="text-primary w-6 h-6" />
            </div>
            <div className="absolute -bottom-1 -right-1 w-4 h-4 bg-green-500 border-2 border-background rounded-full"></div>
          </div>
          <div>
            <h2 className="font-bold text-base tracking-tight">
              Stranger <span className="text-primary/70 ml-1">({partnerGender || '...'})</span>
            </h2>
            <div className="flex items-center space-x-1.5">
              <span className="text-[10px] text-green-500 uppercase font-black tracking-widest">Connected</span>
            </div>
          </div>
        </div>

        <div className="flex items-center space-x-3">
          <button 
            onClick={handleNext}
            className="px-6 py-2 bg-gradient-to-r from-primary to-accent text-white text-sm font-bold rounded-xl transition-all shadow-lg shadow-primary/25 hover:scale-105 active:scale-95"
          >
            Next
          </button>
          <button className="p-2 hover:bg-red-500/10 hover:text-red-500 rounded-xl transition-colors text-foreground/40">
            <Flag className="w-5 h-5" />
          </button>
        </div>
      </header>

      {/* Chat Content */}
      <div className="flex-1 relative overflow-hidden flex flex-col">
        
        {/* Voice Call Overlay */}
        {mode === 'voice' && (
          <div className="absolute inset-0 z-10 glass flex flex-col">
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
          className="flex-1 overflow-y-auto p-6 space-y-6 scroll-smooth"
        >
          {messages.map((msg, i) => (
            <div 
              key={i}
              className={clsx(
                "flex w-full animate-in fade-in slide-in-from-bottom-2 duration-300",
                msg.sender === 'me' ? "justify-end" : "justify-start"
              )}
            >
              <div 
                className={clsx(
                  "chat-bubble max-w-[80%]",
                  msg.sender === 'me' ? "chat-bubble-me" : "chat-bubble-partner"
                )}
              >
                <p className="text-sm font-medium leading-relaxed">{msg.content}</p>
                <span className={clsx(
                  "text-[9px] font-bold uppercase tracking-tighter opacity-50 block mt-1.5",
                  msg.sender === 'me' ? "text-right" : "text-left"
                )}>
                  {new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </span>
              </div>
            </div>
          ))}

          {isPartnerTyping && (
            <div className="flex justify-start animate-in fade-in duration-300">
              <div className="chat-bubble chat-bubble-partner px-4 py-3">
                <div className="flex space-x-1.5">
                  <div className="w-1.5 h-1.5 bg-primary/40 rounded-full animate-bounce"></div>
                  <div className="w-1.5 h-1.5 bg-primary/40 rounded-full animate-bounce [animation-delay:0.2s]"></div>
                  <div className="w-1.5 h-1.5 bg-primary/40 rounded-full animate-bounce [animation-delay:0.4s]"></div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Input Area */}
        <form 
          onSubmit={handleSend}
          className="p-6 glass mt-auto"
        >
          <div className="flex items-center space-x-3 max-w-4xl mx-auto">
            <div className="flex-1 relative group">
              <input 
                type="text"
                value={input}
                onChange={handleInputChange}
                placeholder="Write a message..."
                className="w-full glass bg-foreground/5 border-none rounded-2xl px-5 py-4 text-sm font-medium focus:ring-2 focus:ring-primary/50 outline-none transition-all placeholder:text-foreground/30"
              />
            </div>
            <button 
              type="submit"
              disabled={!input.trim()}
              className="p-4 bg-gradient-to-br from-primary to-accent text-white rounded-2xl hover:scale-105 disabled:opacity-30 disabled:hover:scale-100 transition-all shadow-xl shadow-primary/20 active:scale-95"
            >
              <Send className="w-6 h-6" />
            </button>
          </div>
        </form>
      </div>
    </main>

  );
}