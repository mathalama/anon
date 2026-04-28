import { useChatStore } from '@/store/chatStore';
import { ServerMessage } from '@/types/chat';

class ChatSocket {
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 20;
  private heartbeatInterval: NodeJS.Timeout | null = null;
  private reconnectTimeout: NodeJS.Timeout | null = null;
  private listeners: Map<string, Set<(payload: any) => void>> = new Map();
  private lastRoomId: string | null = null;
  private lastToken: string | null = null;
  private isReconnecting = false;

  connect(roomId: string, token: string) {
    this.lastRoomId = roomId;
    this.lastToken = token;

    // Cancel any pending reconnect
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    if (this.ws) {
      // Remove handlers before closing to prevent reconnect loop
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      this.ws.onopen = null;
      this.ws.close();
      this.ws = null;
    }

    this.isReconnecting = false;
    this.stopHeartbeat();

    const wsUrl = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws';
    const finalUrl = `${wsUrl}?room_id=${roomId}&token=${token}`;
    console.log('Connecting to:', finalUrl);

    this.ws = new WebSocket(finalUrl);

    this.ws.onopen = () => {
      console.log('Connected to chat');
      this.isReconnecting = false;
      this.reconnectAttempts = 0;
      this.startHeartbeat();
    };

    this.ws.onmessage = (e) => {
      try {
        const msg: ServerMessage = JSON.parse(e.data);
        if (msg.type === 'pong') return;
        this.handleMessage(msg);
      } catch (err) {
        console.error('Failed to parse WS message', err);
      }
    };

    this.ws.onclose = (e) => {
      console.log('Disconnected from chat', e.code, e.reason);
      this.stopHeartbeat();

      // Prevent multiple simultaneous reconnect loops
      if (this.isReconnecting) return;

      const status = useChatStore.getState().status;
      if (status !== 'idle' && status !== 'ended' && this.reconnectAttempts < this.maxReconnectAttempts) {
        this.isReconnecting = true;
        this.reconnectAttempts++;
        const delay = Math.min(1000 * Math.pow(1.5, this.reconnectAttempts), 10000);
        console.log(`Attempting reconnect ${this.reconnectAttempts} in ${delay}ms...`);
        this.reconnectTimeout = setTimeout(() => {
          this.isReconnecting = false;
          if (this.lastRoomId && this.lastToken) {
            this.connect(this.lastRoomId, this.lastToken);
          }
        }, delay);
      } else if (status !== 'idle' && status !== 'ended') {
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

    // Notify listeners
    const typeListeners = this.listeners.get(msg.type);
    if (typeListeners) {
      typeListeners.forEach(cb => cb(msg.payload || msg));
    }
  }

  send(content: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'message', content }));
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

  onMessage(type: string, callback: (payload: any) => void) {
    if (!this.listeners.has(type)) {
      this.listeners.set(type, new Set());
    }
    this.listeners.get(type)!.add(callback);
  }

  offMessage(type: string, callback: (payload: any) => void) {
    const typeListeners = this.listeners.get(type);
    if (typeListeners) {
      typeListeners.delete(callback);
    }
  }

  sendRTC(type: string, payload: any) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type, payload }));
    }
  }

  disconnect() {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }
    this.stopHeartbeat();
    this.isReconnecting = false;
    this.reconnectAttempts = 0;
    if (this.ws) {
      this.ws.onclose = null;
      this.ws.onerror = null;
      this.ws.onmessage = null;
      this.ws.onopen = null;
      this.ws.close();
      this.ws = null;
    }
  }

  private startHeartbeat() {
    this.stopHeartbeat();
    this.heartbeatInterval = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);
  }

  private stopHeartbeat() {
    if (this.heartbeatInterval) {
      clearInterval(this.heartbeatInterval);
      this.heartbeatInterval = null;
    }
  }
}

export const chatSocket = new ChatSocket();