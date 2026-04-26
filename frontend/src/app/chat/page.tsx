'use client';

import { useEffect, useRef, useState } from 'react';
import { useRouter } from 'next/navigation';
import { useChat } from '@/hooks/useChat';
import { Send, User, ChevronRight, Flag, X } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export default function ChatPage() {
  const router = useRouter();
  const { status, messages, isPartnerTyping, sendMessage, sendTyping, next } = useChat();
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (status === 'idle') {
      router.push('/');
    } else if (status === 'searching') {
      router.push('/search');
    }
  }, [status, router]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isPartnerTyping]);

  const handleSend = () => {
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

  return (
    <main className="flex flex-col h-screen bg-slate-950">
      {/* Header */}
      <header className="flex items-center justify-between px-6 py-4 bg-slate-900/50 border-b border-slate-800 backdrop-blur-md">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 rounded-full bg-indigo-500/20 flex items-center justify-center">
            <User className="text-indigo-500 w-6 h-6" />
          </div>
          <div>
            <h3 className="font-bold">Stranger</h3>
            <p className="text-xs text-indigo-400">Connected</p>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <button className="p-2 text-slate-400 hover:text-red-500 transition-colors">
            <Flag className="w-5 h-5" />
          </button>
          <button 
            onClick={handleNext}
            className="flex items-center space-x-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 rounded-xl font-bold transition-all active:scale-95"
          >
            <span>Next</span>
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-6 space-y-4">
        <div className="text-center py-8">
          <p className="text-xs text-slate-500 uppercase tracking-widest">Chat started — be respectful</p>
        </div>

        {messages.map((msg, i) => {
          const isMe = msg.sender === 'me';
          return (
            <div key={msg.id} className={cn("flex", isMe ? "justify-end" : "justify-start")}>
              <div className={cn(
                "max-w-[80%] px-4 py-2 rounded-2xl text-sm shadow-lg",
                isMe 
                  ? "bg-indigo-600 text-white rounded-tr-none" 
                  : "bg-slate-800 text-slate-100 rounded-tl-none"
              )}>
                {msg.content}
                <div className={cn("text-[10px] mt-1 opacity-50", isMe ? "text-right" : "text-left")}>
                  {new Date(msg.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </div>
              </div>
            </div>
          );
        })}

        {isPartnerTyping && (
          <div className="flex justify-start">
            <div className="bg-slate-800 px-4 py-2 rounded-2xl rounded-tl-none flex items-center space-x-1">
              <div className="w-1 h-1 bg-slate-400 rounded-full animate-bounce"></div>
              <div className="w-1 h-1 bg-slate-400 rounded-full animate-bounce [animation-delay:0.2s]"></div>
              <div className="w-1 h-1 bg-slate-400 rounded-full animate-bounce [animation-delay:0.4s]"></div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-6 bg-slate-900/50 border-t border-slate-800">
        <div className="flex items-center space-x-4 max-w-4xl mx-auto">
          <input
            type="text"
            value={input}
            onChange={handleInputChange}
            onKeyDown={(e) => e.key === 'Enter' && handleSend()}
            placeholder="Type your message..."
            className="flex-1 bg-slate-800 border-none rounded-2xl px-6 py-3 text-slate-100 placeholder:text-slate-500 focus:ring-2 focus:ring-indigo-500 transition-all outline-none"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim()}
            className="w-12 h-12 rounded-2xl bg-indigo-600 hover:bg-indigo-500 flex items-center justify-center transition-all active:scale-90 disabled:opacity-50 disabled:grayscale"
          >
            <Send className="w-5 h-5 text-white" />
          </button>
        </div>
      </div>

      {/* Status Overlay for Ended */}
      {status === 'ended' && (
        <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm flex items-center justify-center p-6 z-50">
          <div className="bg-slate-900 border border-slate-800 p-8 rounded-3xl max-w-sm w-full text-center space-y-6 shadow-2xl">
            <div className="w-16 h-16 bg-red-500/20 rounded-full flex items-center justify-center mx-auto">
              <X className="text-red-500 w-10 h-10" />
            </div>
            <div className="space-y-2">
              <h2 className="text-2xl font-bold">Chat Ended</h2>
              <p className="text-slate-400">Your partner has disconnected or you moved to the next chat.</p>
            </div>
            <button
              onClick={() => router.push('/search')}
              className="w-full py-4 bg-indigo-600 hover:bg-indigo-500 rounded-2xl font-bold transition-all active:scale-95"
            >
              Find New Partner
            </button>
          </div>
        </div>
      )}
    </main>
  );
}
