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

      <div className="flex flex-col items-center space-y-6">
        <div className="relative">
          <div className={clsx(
            "w-32 h-32 rounded-full bg-gray-200 dark:bg-gray-800 flex items-center justify-center transition-all duration-500",
            callState === 'connected' && "ring-4 ring-green-500/30"
          )}>
            <User className="w-16 h-16 text-gray-400" />
          </div>
          {callState === 'connected' && (
            <div className="absolute bottom-1 right-1 w-6 h-6 bg-green-500 border-4 border-white dark:border-[#1c1c1c] rounded-full animate-pulse"></div>
          )}
        </div>

        <div className="text-center space-y-2">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Stranger ({partnerGender || '...'})
          </h2>
          <p className={clsx(
            "text-sm font-medium uppercase tracking-wider",
            callState === 'connected' ? "text-green-500" : "text-gray-500 animate-pulse"
          )}>
            {callState === 'calling' ? 'Calling...' : callState === 'connected' ? 'Voice Connected' : 'Connecting...'}
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
