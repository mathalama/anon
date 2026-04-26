import { useEffect, useCallback } from 'react';
import { useChatStore } from '@/store/chatStore';
import { chatSocket } from '@/lib/socket';
import { api } from '@/lib/api';

export function useChat() {
  const store = useChatStore();

  const startSearch = useCallback(async (filter = {}) => {
    try {
      store.setStatus('searching');
      await api.search(filter);
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
    let pollInterval: NodeJS.Timeout;

    if (store.status === 'searching') {
      pollInterval = setInterval(async () => {
        try {
          const data = await api.getStatus();
          if (data.status === 'matched' && data.room_id) {
            clearInterval(pollInterval);
            store.setRoomId(data.room_id);
            store.setStatus('matched');
            
            const token = localStorage.getItem('access_token');
            if (token) {
              chatSocket.connect(data.room_id, token);
            }
          }
        } catch (err) {
          console.error('Poll failed', err);
        }
      }, 1000);
    }

    return () => {
      if (pollInterval) clearInterval(pollInterval);
    };
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
