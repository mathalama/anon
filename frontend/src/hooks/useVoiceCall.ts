import { useEffect, useRef, useState, useCallback } from 'react';
import { chatSocket } from '@/lib/socket';
import { useChatStore } from '@/store/chatStore';

export type CallState = 'idle' | 'calling' | 'connected' | 'ended' | 'error';

export function useVoiceCall() {
  const { mode, status, isInitiator } = useChatStore();
  const [callState, setCallState] = useState<CallState>('idle');
  const [isMuted, setIsMuted] = useState(false);
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const isStartingRef = useRef(false);
  const localStreamRef = useRef<MediaStream | null>(null);
  const remoteAudioRef = useRef<HTMLAudioElement | null>(null);

  const cleanup = useCallback(() => {
    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }
    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach(track => track.stop());
      localStreamRef.current = null;
    }
    iceQueue.current = [];
    setCallState('idle');
  }, []);

  const createPeerConnection = useCallback(() => {
    const configuration: RTCConfiguration = {
      iceServers: [
        { urls: 'stun:stun.l.google.com:19302' },
        { urls: 'stun:stun1.l.google.com:19302' },
      ],
      iceCandidatePoolSize: 10,
      bundlePolicy: 'max-bundle',
    };

    const pc = new RTCPeerConnection(configuration);

    pc.onicecandidate = (event) => {
      if (event.candidate) {
        chatSocket.sendRTC('rtc:ice-candidate', event.candidate);
      }
    };

    pc.ontrack = (event) => {
      if (remoteAudioRef.current) {
        remoteAudioRef.current.srcObject = event.streams[0];
      }
      setCallState('connected');
    };

    pc.oniceconnectionstatechange = () => {
      console.log('ICE Connection State:', pc.iceConnectionState);
      if (pc.iceConnectionState === 'failed' || pc.iceConnectionState === 'disconnected') {
        console.log('ICE failed/disconnected, attempting to restart ICE...');
        pc.restartIce();
      }
    };

    pc.onconnectionstatechange = () => {
      console.log('Peer Connection State:', pc.connectionState);
      switch (pc.connectionState) {
        case 'connected':
          setCallState('connected');
          break;
        case 'disconnected':
        case 'failed':
          // Handled by ICE restart, but if it stays failed:
          setCallState('error');
          break;
        case 'closed':
          setCallState('ended');
          break;
      }
    };

    return pc;
  }, []);

  const iceQueue = useRef<RTCIceCandidateInit[]>([]);

  const startCall = useCallback(async () => {
    if (callState === 'connected') return;
    
    try {
      if (pcRef.current) {
        console.log('Cleaning up old connection before startCall');
        pcRef.current.close();
      }
      iceQueue.current = [];
      
      setCallState('calling');
      let stream = localStreamRef.current;
      if (!stream) {
        stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        localStreamRef.current = stream;
      }

      const pc = createPeerConnection();
      pcRef.current = pc;

      stream.getTracks().forEach(track => pc.addTrack(track, stream));

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);

      console.log('Sending RTC Offer');
      chatSocket.sendRTC('rtc:offer', offer);
      chatSocket.sendRTC('call:start', {});
    } catch (err: any) {
      console.error('Failed to start call', err);
      setCallState('error');
    }
  }, [createPeerConnection]);

  const endCall = useCallback(() => {
    chatSocket.sendRTC('call:end', {});
    cleanup();
  }, [cleanup]);

  const toggleMute = useCallback(() => {
    if (localStreamRef.current) {
      const audioTrack = localStreamRef.current.getAudioTracks()[0];
      if (audioTrack) {
        audioTrack.enabled = !audioTrack.enabled;
        setIsMuted(!audioTrack.enabled);
      }
    }
  }, []);

  useEffect(() => {
    let retryInterval: NodeJS.Timeout;
    
    if (mode === 'voice' && status === 'chatting' && isInitiator && !isStartingRef.current && callState === 'idle') {
      isStartingRef.current = true;
      
      const startFast = async () => {
        // Start mic and call in parallel
        const micPromise = (async () => {
          if (!localStreamRef.current) {
            try {
              const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
              localStreamRef.current = stream;
              console.log('Mic ready');
              return stream;
            } catch (e) {
              console.error('Mic failed', e);
            }
          }
          return localStreamRef.current;
        })();

        console.log('Voice session: Fast start initiated');
        await startCall();
        
        retryInterval = setInterval(async () => {
          if (pcRef.current && (pcRef.current.connectionState === 'failed' || pcRef.current.connectionState === 'disconnected')) {
            console.log('Connection dropped: aggressive retry...');
            await startCall();
          }
        }, 3000);
      };
      
      startFast();
    } else if (mode === 'voice' && status === 'chatting' && !isInitiator && !localStreamRef.current) {
      navigator.mediaDevices.getUserMedia({ audio: true }).then(stream => {
        localStreamRef.current = stream;
        console.log('Receiver mic ready early');
      }).catch(e => console.error('Receiver mic failed', e));
    }
    
    return () => {
      clearInterval(retryInterval);
      isStartingRef.current = false;
    };
  }, [mode, status, isInitiator, startCall, callState]);


  useEffect(() => {
  if (mode !== 'voice' || status === 'idle' || status === 'searching') return;

  const handleOffer = async (offer: RTCSessionDescriptionInit) => {
    if (callState === 'connected') {
      console.log('Ignoring offer - already connected');
      return;
    }

    console.log('Received RTC Offer');

    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }

    iceQueue.current = [];

    const pc = createPeerConnection();
    pcRef.current = pc;

    try {
      let stream = localStreamRef.current;
      if (!stream) {
        stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        localStreamRef.current = stream;
      }
      stream.getTracks().forEach(track => pc.addTrack(track, stream!));

      await pc.setRemoteDescription(new RTCSessionDescription(offer));
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);

      console.log('Sending RTC Answer');
      chatSocket.sendRTC('rtc:answer', answer);

      console.log(`Processing ${iceQueue.current.length} queued ICE candidates`);
      for (const candidate of iceQueue.current) {
        try {
          await pc.addIceCandidate(new RTCIceCandidate(candidate));
        } catch (e) {
          console.error('Error adding queued candidate', e);
        }
      }
      iceQueue.current = [];
    } catch (e) {
      console.error('Error handling offer', e);
    }
  };

  const handleAnswer = async (answer: RTCSessionDescriptionInit) => {
    if (!pcRef.current || pcRef.current.signalingState !== 'have-local-offer') {
      console.log('Ignoring answer - wrong state:', pcRef.current?.signalingState);
      return;
    }

    console.log('Received RTC Answer');
    try {
      await pcRef.current.setRemoteDescription(new RTCSessionDescription(answer));
    } catch (e) {
      console.error('Error setting remote answer', e);
    }
  };

  const handleCandidate = async (payload: RTCIceCandidateInit) => {
    const pc = pcRef.current;

    if (!pc || pc.signalingState === 'closed') {
      return;
    }

    if (!pc.remoteDescription) {
      console.log('Queueing ICE candidate');
      iceQueue.current.push(payload);
      return;
    }

    try {
      await pc.addIceCandidate(new RTCIceCandidate(payload));
    } catch (e) {
      console.error('Error adding ice candidate', e);
    }
  };

  const handleCallEnd = () => {
    cleanup();
  };

  chatSocket.onMessage('rtc:offer', handleOffer);
  chatSocket.onMessage('rtc:answer', handleAnswer);
  chatSocket.onMessage('rtc:ice-candidate', handleCandidate);
  chatSocket.onMessage('call:end', handleCallEnd);

  return () => {
    chatSocket.offMessage('rtc:offer', handleOffer);
    chatSocket.offMessage('rtc:answer', handleAnswer);
    chatSocket.offMessage('rtc:ice-candidate', handleCandidate);
    chatSocket.offMessage('call:end', handleCallEnd);
  };
}, [mode, status, createPeerConnection, cleanup, callState]);

  // Handle room end
  useEffect(() => {
    if (status === 'ended') {
      cleanup();
    }
  }, [status, cleanup]);

  return {
    startCall,
    endCall,
    isMuted,
    toggleMute,
    callState,
    remoteAudioRef,
  };
}
