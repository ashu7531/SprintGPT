import { useState, useRef, useEffect } from 'react';

export default function InputBar({ onSend, disabled, externalValue, onExternalValueUsed }) {
  const [message, setMessage] = useState('');
  const inputRef = useRef(null);

  // When a suggestion is clicked, fill the input
  useEffect(() => {
    if (externalValue) {
      setMessage(externalValue);
      onExternalValueUsed && onExternalValueUsed();
      inputRef.current?.focus();
    }
  }, [externalValue]);

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
    <div style={{ flexShrink: 0, width: '100%', display: 'flex', justifyContent: 'center', padding: '16px 32px 28px', backgroundColor: '#171717' }}>
      <div style={{ width: '100%', maxWidth: '720px', position: 'relative' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px', border: '1px solid #525252', borderRadius: '16px', backgroundColor: '#262626', padding: '14px 16px' }}>
          <textarea
            ref={inputRef}
            id="chat-input"
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Message SprintGPT..."
            rows={1}
            disabled={disabled}
            style={{ flex: 1, background: 'transparent', color: '#f5f5f5', resize: 'none', outline: 'none', fontSize: '15px', lineHeight: '1.5', maxHeight: '144px', border: 'none', padding: 0, fontFamily: 'inherit' }}
          />
          <button
            type="submit"
            disabled={!message.trim() || disabled}
            onClick={handleSubmit}
            style={{ flexShrink: 0, width: '36px', height: '36px', borderRadius: '10px', backgroundColor: !message.trim() || disabled ? '#404040' : '#10b981', color: !message.trim() || disabled ? '#737373' : 'white', display: 'flex', alignItems: 'center', justifyContent: 'center', border: 'none', cursor: !message.trim() || disabled ? 'not-allowed' : 'pointer' }}
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" style={{ width: '16px', height: '16px' }}>
              <path fillRule="evenodd" d="M10 17a.75.75 0 0 1-.75-.75V5.612L5.29 9.77a.75.75 0 0 1-1.08-1.04l5.25-5.5a.75.75 0 0 1 1.08 0l5.25 5.5a.75.75 0 1 1-1.08 1.04l-3.96-4.158V16.25A.75.75 0 0 1 10 17Z" clipRule="evenodd" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  );
}
