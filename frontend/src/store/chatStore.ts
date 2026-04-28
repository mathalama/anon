import { create } from 'zustand';
import { ChatState, Message, ChatStatus } from '@/types/chat';

interface ChatStore extends ChatState {
  selectedMode: 'text' | 'voice' | '';
  myGender: 'male' | 'female' | '';
  selectedGender: 'any' | 'male' | 'female' | '';
  mode: 'text' | 'voice' | null;
  isInitiator: boolean;
  partnerGender: string;
  setStatus: (status: ChatStatus) => void;
  setRoomId: (roomId: string | null) => void;
  setMode: (mode: 'text' | 'voice' | null) => void;
  setIsInitiator: (isInitiator: boolean) => void;
  setSelectedMode: (mode: 'text' | 'voice' | '') => void;
  setSelectedGender: (gender: 'any' | 'male' | 'female' | '') => void;
  setMyGender: (gender: 'male' | 'female' | '') => void;
  setPartnerGender: (gender: string) => void;
  addMessage: (message: Message) => void;
  setPartnerTyping: (isTyping: boolean) => void;
  setEndReason: (reason: 'next' | 'disconnect' | 'ban' | null) => void;
  setMatchData: (data: {
    roomId: string;
    mode: 'text' | 'voice';
    isInitiator: boolean;
    partnerGender: string;
  }) => void;
  reset: () => void;
}

export const useChatStore = create<ChatStore>((set) => ({
  status: 'idle',
  roomId: null,
  messages: [],
  isPartnerTyping: false,
  endReason: null,
  selectedMode: '',
  myGender: '',
  selectedGender: '',
  mode: null,
  isInitiator: false,
  partnerGender: '',

  setStatus: (status) => set({ status }),
  setRoomId: (roomId) => set({ roomId }),
  setMode: (mode) => set({ mode }),
  setIsInitiator: (isInitiator) => set({ isInitiator }),
  setSelectedMode: (mode) => set({ selectedMode: mode }),
  setSelectedGender: (gender) => set({ selectedGender: gender }),
  setMyGender: (gender) => set({ myGender: gender }),
  setPartnerGender: (gender) => set({ partnerGender: gender }),
  addMessage: (message) => set((state) => ({ messages: [...state.messages, message] })),
  setPartnerTyping: (isTyping) => set({ isPartnerTyping: isTyping }),
  setEndReason: (reason) => set({ endReason: reason }),
  setMatchData: (data) => set({
    roomId: data.roomId,
    mode: data.mode,
    isInitiator: data.isInitiator,
    partnerGender: data.partnerGender,
    messages: [],
    isPartnerTyping: false,
    endReason: null,
    status: 'matched',
  }),
  reset: () => set({ 
    status: 'idle', roomId: null, messages: [], 
    isPartnerTyping: false, endReason: null, mode: null, 
    isInitiator: false, selectedGender: '', partnerGender: '',
    selectedMode: '', myGender: ''
  }),
}));


