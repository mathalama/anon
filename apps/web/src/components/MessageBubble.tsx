import { memo } from 'react';
import { clsx } from 'clsx';
import { Message } from '@/types/chat';

export const MessageBubble = memo(function MessageBubble({ msg }: { msg: Message }) {
  return (
    <div
      className={clsx(
        'flex w-full',
        msg.sender === 'me' ? 'justify-end' : 'justify-start',
      )}
    >
      <div
        className={clsx(
          'chat-bubble',
          msg.sender === 'me' ? 'chat-bubble-me' : 'chat-bubble-partner',
        )}
      >
        {msg.content && <p>{msg.content}</p>}
        <span
          className={clsx(
            'text-[10px] font-medium uppercase tracking-widest mt-2 block opacity-40',
            msg.sender === 'me' ? 'text-right' : 'text-left',
          )}
        >
          {new Date(typeof msg.timestamp === 'number' && msg.timestamp > 1e12 ? msg.timestamp : (msg.timestamp || 0) * 1000).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
        </span>
      </div>
    </div>
  );
});
