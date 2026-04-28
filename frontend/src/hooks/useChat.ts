import { useEffect, useCallback } from 'react';
import { useChatStore } from '@/store/chatStore';
import { chatSocket } from '@/lib/socket';
import { api } from '@/lib/api';

export function useChat() {
  const store = useChatStore();

  const startSearch = useCallback(async (filter: any = {}) => {
    try {
      store.setStatus('searching');
      await api.search({
        my_gender: store.myGender,
        gender: store.selectedGender,
        mode: store.selectedMode
      });
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
    if (store.status === 'searching') {
      let token = localStorage.getItem('access_token') || '';
      if (token === 'undefined' || token === 'null') token = '';
      
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
            console.log('useChat: Match confirmed! RoomID:', data.room_id);
            store.setMode(data.mode || 'text');
            store.setIsInitiator(!!data.is_initiator);
            store.setPartnerGender(data.partner_gender || 'unknown');
            store.setRoomId(data.room_id);
            store.setStatus('matched');
            chatSocket.connect(data.room_id, token);
            eventSource.close();
          }
        } catch (err) {
          console.error('useChat: SSE data parse failed', err);
        }
      };

      eventSource.onerror = (err) => {
        console.error('useChat: SSE Error details:', err);
        eventSource.close();
      };

      return () => {
        eventSource.close();
      };
    }
  }, [store.status, store.setRoomId, store.setStatus]);

  return {
    ...store,
    startSearch,
    cancelSearch,
    sendMessage: chatSocket.send.bind(chatSocket),
    sendTyping: chatSocket.sendTyping.bind(chatSocket),
    next: chatSocket.next.bind(chatSocket),
  };
}
