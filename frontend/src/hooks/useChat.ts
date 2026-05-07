import { useEffect, useCallback, useRef } from 'react';
import { useChatStore } from '@/store/chatStore';
import { chatSocket } from '@/lib/socket';
import { api } from '@/lib/api';

export function useChat() {
  const store = useChatStore();
  // Track if we've already connected for this match session
  const connectedRoomRef = useRef<string | null>(null);
  const retryTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const startSearch = useCallback(async () => {
    try {
      await api.search({
        my_gender: store.myGender,
        gender: store.selectedGender,
        mode: store.selectedMode,
      });
      
      store.setStatus('searching');
    } catch (err) {
      console.error('Search failed', err);
      store.setStatus('idle');
    }
  }, [store]);

  const cancelSearch = useCallback(async () => {
    try {
      await api.cancelSearch();
      store.setStatus('idle');
    } catch (err) {
      console.error('Cancel failed', err);
    }
  }, [store]);

  useEffect(() => {
    if (store.status !== 'searching') return;

    let token = localStorage.getItem('access_token') || '';
    if (token === 'undefined' || token === 'null') token = '';
    let isClosedByEffect = false;

    const sseUrl = api.getMatchSSEUrl(token);
    console.log('useChat: Initiating SSE connection to:', sseUrl);

    const eventSource = new EventSource(sseUrl);

    eventSource.onopen = () => {
      console.log('useChat: SSE connection opened');
    };

    eventSource.onmessage = (event) => {
      console.log('useChat: SSE message received:', event.data);
      try {
        const data = JSON.parse(event.data);
        if (data.status === 'matched' && data.room_id) {
          // Prevent double-connecting to same room
          if (connectedRoomRef.current === data.room_id) {
            console.log('useChat: Already connected to room', data.room_id, '- ignoring duplicate SSE');
            eventSource.close();
            return;
          }

          console.log('useChat: Match confirmed! RoomID:', data.room_id);
          connectedRoomRef.current = data.room_id;

          store.setMatchData({
            roomId: data.room_id,
            mode: data.mode || 'text',
            isInitiator: !!data.is_initiator,
            partnerGender: data.partner_gender || 'unknown',
            partnerUserId: data.partner_user_id || '',
          });

          const t = localStorage.getItem('access_token') || '';
          chatSocket.connect(data.room_id, t);
          eventSource.close();
        }
      } catch (err) {
        console.error('useChat: SSE data parse failed', err);
      }
    };

    eventSource.onerror = (err) => {
      console.error('useChat: SSE Error details:', err);
      eventSource.close();
      if (isClosedByEffect || store.status !== 'searching') {
        return;
      }

      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
      retryTimeoutRef.current = setTimeout(() => {
        retryTimeoutRef.current = null;
        store.setStatus('idle');
        store.setStatus('searching');
      }, 1000);
    };

    return () => {
      isClosedByEffect = true;
      eventSource.close();
    };
  }, [store, store.status]); // Only re-run when status changes

  // Reset connectedRoomRef when going back to idle/searching
  useEffect(() => {
    if (store.status === 'idle' || store.status === 'searching') {
      connectedRoomRef.current = null;
    }
  }, [store.status]);

  useEffect(() => {
    return () => {
      if (retryTimeoutRef.current) {
        clearTimeout(retryTimeoutRef.current);
      }
    };
  }, []);

  return {
    ...store,
    startSearch,
    cancelSearch,
    sendMessage: chatSocket.send.bind(chatSocket),
    sendTyping: chatSocket.sendTyping.bind(chatSocket),
    next: chatSocket.next.bind(chatSocket),
  };
}
