import { useState, useEffect, useRef } from 'react';

export default function ChatWindow({ messages, isLoading }) {
  const messagesEndRef = useRef(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, isLoading]);

  return (
    <div className="flex-1 overflow-y-auto p-4 space-y-4">
      {messages.length === 0 && !isLoading && (
        <div className="flex flex-col items-center justify-center h-full text-center px-4">
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center mb-6 logo-glow">
            <span className="text-3xl">⚡</span>
          </div>
          <h2 className="text-2xl font-bold text-surface-100 mb-2">
            Welcome to SprintGPT
          </h2>
          <p className="text-surface-400 max-w-md mb-8">
            Your AI-powered Azure DevOps assistant. Ask me about tasks, bugs, sprints, and more.
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-lg w-full">
            {[
              { icon: '🔍', text: 'What is the status of task 123?' },
              { icon: '👤', text: 'What is Rahul working on?' },
              { icon: '🐛', text: 'Show active bugs' },
              { icon: '📊', text: 'Sprint summary' },
            ].map((example, i) => (
              <div
                key={i}
                className="p-3 rounded-xl bg-surface-800/50 border border-surface-700/50 text-sm text-surface-300 hover:bg-surface-800 hover:border-primary-500/30 transition-all duration-200 cursor-default"
              >
                <span className="mr-2">{example.icon}</span>
                {example.text}
              </div>
            ))}
          </div>
        </div>
      )}

      {messages.map((msg, index) => (
        <MessageBubble key={index} message={msg} />
      ))}

      {isLoading && <TypingIndicator />}

      <div ref={messagesEndRef} />
    </div>
  );
}

function MessageBubble({ message }) {
  const isUser = message.role === 'user';

  return (
    <div className={`message-enter flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div className={`flex items-start gap-3 max-w-[80%] ${isUser ? 'flex-row-reverse' : ''}`}>
        {/* Avatar */}
        <div
          className={`w-8 h-8 rounded-lg flex items-center justify-center shrink-0 text-sm font-semibold ${
            isUser
              ? 'bg-primary-600 text-white'
              : 'bg-gradient-to-br from-primary-500 to-purple-600 text-white'
          }`}
        >
          {isUser ? 'U' : '⚡'}
        </div>

        {/* Message bubble */}
        <div
          className={`rounded-2xl px-4 py-3 text-sm leading-relaxed ${
            isUser
              ? 'bg-primary-600 text-white rounded-tr-sm'
              : 'bg-surface-800 text-surface-200 border border-surface-700/50 rounded-tl-sm'
          }`}
        >
          <div className="message-content whitespace-pre-wrap text-sm leading-relaxed text-surface-200">
            {message.content || ''}
          </div>

          {message.intent && !isUser && (
            <div className="mt-2 pt-2 border-t border-surface-700/30">
              <span className="text-xs px-2 py-0.5 rounded-full bg-primary-500/20 text-primary-300 font-medium">
                {message.intent}
              </span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

function TypingIndicator() {
  return (
    <div className="message-enter flex justify-start">
      <div className="flex items-start gap-3">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary-500 to-purple-600 flex items-center justify-center text-sm text-white">
          ⚡
        </div>
        <div className="bg-surface-800 border border-surface-700/50 rounded-2xl rounded-tl-sm px-4 py-3">
          <div className="flex gap-1.5">
            <div className="typing-dot w-2 h-2 rounded-full bg-primary-400" />
            <div className="typing-dot w-2 h-2 rounded-full bg-primary-400" />
            <div className="typing-dot w-2 h-2 rounded-full bg-primary-400" />
          </div>
        </div>
      </div>
    </div>
  );
}
