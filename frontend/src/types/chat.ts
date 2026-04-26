export type ChatStatus = 'idle' | 'searching' | 'matched' | 'chatting' | 'ended';

export interface Message {
  id: string;
  sender: string; // 'me' | 'partner'
  content: string;
  timestamp: number;
}

export interface ServerMessage {
  type: 'partner_connected' | 'message' | 'partner_disconnected' | 'partner_typing' | 'match_found' | 'error';
  room_id?: string;
  content?: string;
  sender?: string;
  is_typing?: boolean;
  timestamp?: number;
}

export type ChatState = {
  status: ChatStatus;
  roomId: string | null;
  messages: Message[];
  isPartnerTyping: boolean;
  endReason: 'next' | 'disconnect' | 'ban' | null;
};
