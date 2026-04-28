'use client';

import { Mic, MicOff, PhoneOff, User } from 'lucide-react';
import { clsx } from 'clsx';

interface VoiceCallUIProps {
  callState: string;
  isMuted: boolean;
  toggleMute: () => void;
  endCall: () => void;
  remoteAudioRef: React.RefObject<HTMLAudioElement | null>;
  partnerGender: string;
}

export function VoiceCallUI({ 
  callState, 
  isMuted, 
  toggleMute, 
  endCall, 
  remoteAudioRef,
  partnerGender 
}: VoiceCallUIProps) {
  return (
    <div className="flex flex-col items-center justify-center h-full p-8 space-y-12">
      {/* Remote Audio element (hidden) */}
      <audio ref={remoteAudioRef} autoPlay />

      <div className="flex flex-col items-center space-y-8">
        <div className="relative">
          {callState === 'connected' && (
            <>
              <div className="ripple"></div>
              <div className="ripple"></div>
              <div className="ripple"></div>
            </>
          )}
          <div className={clsx(
            "relative w-36 h-36 rounded-full bg-secondary flex items-center justify-center transition-all duration-500 glass border-2 border-primary/20",
            callState === 'connected' && "scale-110 shadow-2xl shadow-primary/40 border-primary/50"
          )}>
            <User className={clsx(
              "w-20 h-20 transition-colors duration-500",
              callState === 'connected' ? "text-primary" : "text-foreground/20"
            )} />
          </div>
          {callState === 'connected' && (
            <div className="absolute -bottom-1 -right-1 w-8 h-8 bg-green-500 border-4 border-background rounded-full z-10"></div>
          )}
        </div>

        <div className="text-center space-y-3">
          <h2 className="text-3xl font-black tracking-tight text-foreground">
            Stranger <span className="text-primary/70">({partnerGender || '...'})</span>
          </h2>
          <p className={clsx(
            "text-xs font-black uppercase tracking-[0.2em]",
            callState === 'connected' ? "text-green-500" : "text-foreground/30 animate-pulse"
          )}>
            {callState === 'calling' ? 'Establishing Connection...' : callState === 'connected' ? 'Voice Active' : 'Connecting...'}
          </p>
        </div>
      </div>

      <div className="flex items-center space-x-8">
        <button
          onClick={toggleMute}
          className={clsx(
            "p-5 rounded-full transition-all shadow-sm",
            isMuted 
              ? "bg-red-500 text-white" 
              : "bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-300 hover:bg-gray-200"
          )}
        >
          {isMuted ? <MicOff className="w-7 h-7" /> : <Mic className="w-7 h-7" />}
        </button>

        <button
          onClick={endCall}
          className="p-5 bg-red-600 hover:bg-red-700 text-white rounded-full transition-all shadow-lg active:scale-95"
        >
          <PhoneOff className="w-7 h-7" />
        </button>
      </div>
    </div>
  );
}
