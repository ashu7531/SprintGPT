import { useState, useRef } from 'react';

export default function InputBar({ onSend, disabled }) {
  const [message, setMessage] = useState('');
  const inputRef = useRef(null);

  const handleSubmit = (e) => {
    e.preventDefault();
    const trimmed = message.trim();
    if (!trimmed || disabled) return;

    onSend(trimmed);
    setMessage('');
    inputRef.current?.focus();
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e);
    }
  };

  return (
    <div className="border-t border-surface-800 bg-surface-950/80 backdrop-blur-sm p-4">
      <form onSubmit={handleSubmit} className="max-w-4xl mx-auto">
        <div className="flex items-end gap-3 bg-surface-800 rounded-2xl border border-surface-700/50 focus-within:border-primary-500/50 transition-colors duration-200 p-2">
          <textarea
            ref={inputRef}
            id="chat-input"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Ask about tasks, bugs, sprints..."
            rows={1}
            disabled={disabled}
            className="flex-1 bg-transparent text-surface-100 placeholder:text-surface-500 resize-none outline-none px-2 py-1.5 text-sm max-h-32 disabled:opacity-50"
            style={{ lineHeight: '1.5' }}
          />
          <button
            type="submit"
            disabled={!message.trim() || disabled}
            className="shrink-0 w-9 h-9 rounded-xl bg-primary-600 hover:bg-primary-500 disabled:bg-surface-700 disabled:text-surface-500 text-white flex items-center justify-center transition-all duration-200 disabled:cursor-not-allowed"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 20 20"
              fill="currentColor"
              className="w-4 h-4"
            >
              <path d="M3.105 2.289a.75.75 0 00-.826.95l1.414 4.925A1.5 1.5 0 005.135 9.25h6.115a.75.75 0 010 1.5H5.135a1.5 1.5 0 00-1.442 1.086l-1.414 4.926a.75.75 0 00.826.95 28.896 28.896 0 0015.293-7.154.75.75 0 000-1.115A28.897 28.897 0 003.105 2.289z" />
            </svg>
          </button>
        </div>
        <p className="text-xs text-surface-600 text-center mt-2">
          SprintGPT queries your Azure DevOps project in real-time
        </p>
      </form>
    </div>
  );
}
