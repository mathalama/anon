import { useChatStore } from '@/store/chatStore';
import { ServerMessage } from '@/types/chat';

class ChatSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 3;

  connect(roomId: string, token: string) {
    if (this.ws) {
      this.ws.close();
    }

    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';
    this.ws = new WebSocket(`${wsUrl}?room_id=${roomId}&token=${token}`);

    this.ws.onopen = () => {
      console.log('Connected to chat');
      this.reconnectAttempts = 0;
    };

    this.ws.onmessage = (e) => {
      try {
        const msg: ServerMessage = JSON.parse(e.data);
        this.handleMessage(msg);
      } catch (err) {
        console.error('Failed to parse WS message', err);
      }
    };

    this.ws.onclose = (e) => {
      console.log('Disconnected from chat', e.reason);
      if (!e.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.reconnectAttempts++;
        setTimeout(() => this.connect(roomId, token), 1000 * Math.pow(2, this.reconnectAttempts));
      } else {
        useChatStore.getState().setStatus('ended');
        useChatStore.getState().setEndReason('disconnect');
      }
    };

    this.ws.onerror = (err) => {
      console.error('WS Error', err);
    };
  }

  private handleMessage(msg: ServerMessage) {
    const store = useChatStore.getState();

    switch (msg.type) {
      case 'partner_connected':
        store.setStatus('chatting');
        break;
      case 'message':
        if (msg.content) {
          store.addMessage({
            id: Math.random().toString(36).substring(7),
            sender: msg.sender || 'partner',
            content: msg.content,
            timestamp: msg.timestamp || Date.now(),
          });
        }
        break;
      case 'partner_disconnected':
        store.setStatus('ended');
        store.setEndReason('disconnect');
        break;
      case 'partner_typing':
        store.setPartnerTyping(!!msg.is_typing);
        break;
    }
  }

  send(content: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'message', content }));
      // Removed local addition to prevent duplication. 
      // Server will broadcast it back to us.
    }
  }

  sendTyping(isTyping: boolean) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'typing', is_typing: isTyping }));
    }
  }

  next() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'next' }));
    }
    this.disconnect();
    useChatStore.getState().setEndReason('next');
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }
}

export const chatSocket = new ChatSocket();
