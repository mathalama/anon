import { useState } from 'react';
import { Send } from 'lucide-react';

interface ChatInputProps {
  onSendMessage: (content: string) => void;
  onTyping: (isTyping: boolean) => void;
}

export function ChatInput({ onSendMessage, onTyping }: ChatInputProps) {
  const [input, setInput] = useState('');

  const handleSend = (e?: React.FormEvent) => {
    e?.preventDefault();
    if (input.trim()) {
      onSendMessage(input.trim());
      setInput('');
      onTyping(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value);
    onTyping(e.target.value.length > 0);
  };

  return (
    <form onSubmit={handleSend} className="p-6 border-t border-white/5 bg-[#050505] mt-auto">
      <div className="flex items-center gap-3 max-w-5xl mx-auto">
        <input
          type="text"
          value={input}
          onChange={handleInputChange}
          placeholder="Напишите сообщение..."
          className="sleek-input flex-1"
        />
        <button
          type="submit"
          disabled={!input.trim()}
          className="sleek-button p-4 rounded-xl flex-shrink-0"
        >
          <Send className="w-6 h-6" />
        </button>
      </div>
    </form>
  );
}
