import { create } from 'zustand';
import { ChatState, Message, ChatStatus } from '@/types/chat';

interface ChatStore extends ChatState {
  setStatus: (status: ChatStatus) => void;
  setRoomId: (roomId: string | null) => void;
  addMessage: (message: Message) => void;
  setPartnerTyping: (isTyping: boolean) => void;
  setEndReason: (reason: 'next' | 'disconnect' | 'ban' | null) => void;
  reset: () => void;
}

export const useChatStore = create<ChatStore>((set) => ({
  status: 'idle',
  roomId: null,
  messages: [],
  isPartnerTyping: false,
  endReason: null,

  setStatus: (status) => set({ status }),
  setRoomId: (roomId) => set({ roomId }),
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  setPartnerTyping: (isTyping) => set({ isPartnerTyping: isTyping }),
  setEndReason: (reason) => set({ endReason: reason }),
  reset: () => set({ status: 'idle', roomId: null, messages: [], isPartnerTyping: false, endReason: null }),
}));
