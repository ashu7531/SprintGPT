import { useState, useEffect } from 'react';
import ChatWindow from './components/Chat/ChatWindow';
import InputBar from './components/Chat/InputBar';
import ConfigPanel from './components/Settings/ConfigPanel';
import { sendMessage, checkHealth } from './services/api';

export default function App() {
  const [messages, setMessages] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showConfig, setShowConfig] = useState(false);
  const [config, setConfig] = useState({ organization: '', project: '', pat: '' });
  const [serverStatus, setServerStatus] = useState('checking');

  // Load config from localStorage on mount
  useEffect(() => {
    const saved = localStorage.getItem('sprintgpt-config');
    if (saved) {
      try {
        setConfig(JSON.parse(saved));
      } catch (e) {
        // ignore bad data
      }
    }
  }, []);

  // Check backend health on mount
  useEffect(() => {
    checkHealth()
      .then(() => setServerStatus('online'))
      .catch(() => setServerStatus('offline'));
  }, []);

  const isConfigured = config.organization && config.project && config.pat;

  const handleSend = async (message) => {
    // Add user message
    const userMsg = { role: 'user', content: message };
    setMessages((prev) => [...prev, userMsg]);
    setIsLoading(true);

    try {
      if (!isConfigured) {
        // If not configured, show a helpful message
        setMessages((prev) => [
          ...prev,
          {
            role: 'assistant',
            content:
              '⚠️ **Azure DevOps not configured!**\n\nClick the ⚙️ settings button in the top-right to enter your organization, project, and PAT.\n\nI need these to fetch your data from Azure DevOps.',
            intent: 'system',
          },
        ]);
        setIsLoading(false);
        return;
      }

      const response = await sendMessage(message, config);

      setMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          content: response.response,
          intent: response.intent,
          data: response.data,
        },
      ]);
    } catch (error) {
      setMessages((prev) => [
        ...prev,
        {
          role: 'assistant',
          content: `❌ **Error:** ${error.message}\n\nMake sure your backend server is running and your Azure DevOps credentials are correct.`,
          intent: 'error',
        },
      ]);
    }

    setIsLoading(false);
  };

  return (
    <div className="h-full flex flex-col bg-surface-950">
      {/* Header */}
      <header className="shrink-0 flex items-center justify-between px-5 py-3 border-b border-surface-800 bg-surface-900/80 backdrop-blur-sm">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center logo-glow">
            <span className="text-lg">⚡</span>
          </div>
          <div>
            <h1 className="text-base font-semibold text-surface-100">SprintGPT</h1>
            <p className="text-xs text-surface-500">Azure DevOps AI Assistant</p>
          </div>
        </div>

        <div className="flex items-center gap-3">
          {/* Server status */}
          <div className="flex items-center gap-1.5 text-xs">
            <div
              className={`w-2 h-2 rounded-full ${
                serverStatus === 'online'
                  ? 'bg-green-400'
                  : serverStatus === 'offline'
                  ? 'bg-red-400'
                  : 'bg-yellow-400 animate-pulse'
              }`}
            />
            <span className="text-surface-500">
              {serverStatus === 'online'
                ? 'Backend online'
                : serverStatus === 'offline'
                ? 'Backend offline'
                : 'Checking...'}
            </span>
          </div>

          {/* Config indicator */}
          {isConfigured && (
            <div className="hidden sm:flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-primary-500/10 border border-primary-500/20">
              <span className="text-xs text-primary-300 font-medium">{config.organization}/{config.project}</span>
            </div>
          )}

          {/* Settings button */}
          <button
            id="settings-button"
            onClick={() => setShowConfig(true)}
            className="w-9 h-9 rounded-xl hover:bg-surface-800 flex items-center justify-center text-surface-400 hover:text-surface-200 transition-colors"
            title="Settings"
          >
            ⚙️
          </button>
        </div>
      </header>

      {/* Chat area */}
      <ChatWindow messages={messages} isLoading={isLoading} />

      {/* Input */}
      <InputBar onSend={handleSend} disabled={isLoading} />

      {/* Config modal */}
      {showConfig && (
        <ConfigPanel config={config} setConfig={setConfig} onClose={() => setShowConfig(false)} />
      )}
    </div>
  );
}
