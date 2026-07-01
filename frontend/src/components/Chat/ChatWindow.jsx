import { useEffect, useRef } from 'react';
import { marked } from 'marked';
import { DataCard } from './ChatCards';

marked.setOptions({ breaks: true, gfm: true });

const SparklesIcon = ({ size = 16, color = '#34d399' }) => (
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill={color} style={{ width: `${size}px`, height: `${size}px` }}>
    <path d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z" />
  </svg>
);

export default function ChatWindow({ messages, isLoading, onSuggestionClick }) {
  const messagesEndRef = useRef(null);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isLoading]);

  if (messages.length === 0 && !isLoading) {
    return (
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '32px' }}>
        <EmptyState onSuggestionClick={onSuggestionClick} />
      </div>
    );
  }

  return (
    <div style={{ flex: 1, overflowY: 'auto' }}>
      <div style={{ maxWidth: '760px', margin: '0 auto', padding: '40px 48px' }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '32px' }}>
          {messages.map((msg, index) => (
            <MessageBubble key={index} message={msg} />
          ))}
          {isLoading && <TypingIndicator />}
          <div ref={messagesEndRef} />
        </div>
      </div>
    </div>
  );
}

function EmptyState({ onSuggestionClick }) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', textAlign: 'center' }}>
      <div style={{ width: '48px', height: '48px', borderRadius: '12px', backgroundColor: '#10b981', display: 'flex', alignItems: 'center', justifyContent: 'center', marginBottom: '20px' }}>
        <SparklesIcon size={24} color="white" />
      </div>
      <h2 style={{ fontSize: '24px', fontWeight: 600, color: '#fafafa', marginBottom: '8px' }}>What can I help with?</h2>
      <p style={{ color: '#a3a3a3', fontSize: '15px', maxWidth: '420px', marginBottom: '32px', lineHeight: 1.6 }}>Ask me about your Azure DevOps tasks, bugs, sprints, or team workload.</p>
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px', width: '100%', maxWidth: '480px' }}>
        {[
          { icon: '🔍', text: 'Status of task #123' },
          { icon: '👤', text: 'What is Rahul working on?' },
          { icon: '🐛', text: 'Show active bugs' },
          { icon: '📊', text: 'Sprint summary' },
        ].map((example, i) => (
          <div
            key={i}
            onClick={() => onSuggestionClick && onSuggestionClick(example.text)}
            style={{ padding: '16px', borderRadius: '12px', border: '1px solid #404040', fontSize: '14px', color: '#d4d4d4', textAlign: 'center', cursor: 'pointer' }}
          >
            <span style={{ display: 'block', fontSize: '18px', marginBottom: '6px' }}>{example.icon}</span>
            <span>{example.text}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function MessageBubble({ message }) {
  const isUser = message.role === 'user';

  if (isUser) {
    return (
      <div className="message-enter" style={{ display: 'flex', justifyContent: 'flex-end' }}>
        <div style={{ maxWidth: '70%', backgroundColor: '#059669', color: 'white', borderRadius: '20px 20px 4px 20px', padding: '14px 20px', fontSize: '15px', lineHeight: 1.6 }}>
          <div style={{ whiteSpace: 'pre-wrap' }}>{message.content || ''}</div>
        </div>
      </div>
    );
  }

  const hasData = message.data && (typeof message.data === 'object');
  const intent = message.intent;

  return (
    <div className="message-enter" style={{ display: 'flex', alignItems: 'flex-start', gap: '12px' }}>
      <div style={{ width: '32px', height: '32px', borderRadius: '8px', backgroundColor: '#262626', border: '1px solid #404040', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
        <SparklesIcon size={16} color="#34d399" />
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        {hasData && <DataCard intent={intent} data={message.data} />}

        <div style={{ backgroundColor: '#1f1f1f', border: '1px solid #2e2e2e', borderRadius: '16px 16px 16px 4px', padding: '16px 20px', marginTop: hasData ? '10px' : '0' }}>
          <div style={{ fontSize: '11px', color: '#737373', marginBottom: '8px', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.5px' }}>
            {hasData ? 'AI Summary' : 'Response'}
          </div>
          <div
            className="message-content"
            style={{ fontSize: '14px', lineHeight: 1.8, color: '#e5e5e5' }}
            dangerouslySetInnerHTML={{ __html: marked.parse(message.content || '') }}
          />
        </div>

        {intent && (
          <div style={{ marginTop: '8px' }}>
            <span style={{ fontSize: '11px', padding: '3px 8px', borderRadius: '4px', backgroundColor: '#262626', border: '1px solid #333', color: '#737373', fontWeight: 500 }}>
              {intent}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}

function TypingIndicator() {
  return (
    <div className="message-enter" style={{ display: 'flex', alignItems: 'flex-start', gap: '12px' }}>
      <div style={{ width: '32px', height: '32px', borderRadius: '8px', backgroundColor: '#262626', border: '1px solid #404040', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
        <SparklesIcon size={16} color="#34d399" />
      </div>
      <div style={{ display: 'flex', gap: '6px', paddingTop: '12px' }}>
        <div className="typing-dot" style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: '#737373' }} />
        <div className="typing-dot" style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: '#737373' }} />
        <div className="typing-dot" style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: '#737373' }} />
      </div>
    </div>
  );
}
