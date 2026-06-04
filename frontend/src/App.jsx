import { useState, useEffect } from 'react';
import ChatWindow from './components/Chat/ChatWindow';
import InputBar from './components/Chat/InputBar';
import ConfigPanel from './components/Settings/ConfigPanel';
import Auth from './components/Auth/Auth';
import { sendMessage, checkHealth, loadUserConfig } from './services/api';
import { supabase } from './utils/supabase';

export default function App() {
  const [session, setSession] = useState(null);
  const [messages, setMessages] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showConfig, setShowConfig] = useState(false);
  const [config, setConfig] = useState({ organization: '', project: '', pat: '' });
  const [serverStatus, setServerStatus] = useState('checking');

  // Load config from database when user session is loaded
  useEffect(() => {
    if (session?.access_token) {
      loadUserConfig(session.access_token)
        .then((savedConfig) => {
          if (savedConfig && savedConfig.organization) {
            setConfig(savedConfig);
          } else {
            // Fallback to localStorage if database does not contain config yet
            const saved = localStorage.getItem('sprintgpt-config');
            if (saved) {
              try {
                setConfig(JSON.parse(saved));
              } catch (e) {
                // ignore bad data
              }
            }
          }
        })
        .catch((err) => {
          console.error('Failed to load user config from database:', err);
        });
    }
  }, [session]);

  // Listen to Supabase Auth changes
  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => {
      setSession(session);
    });

    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange((_event, session) => {
      setSession(session);
    });

    return () => subscription.unsubscribe();
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
      // Send message including the active session token
      const response = await sendMessage(message, config, session?.access_token);

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

  // If there is no active session, show the Login/Signup screen
  if (!session) {
    return <Auth onAuthSuccess={(sess) => setSession(sess)} />;
  }

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

          {/* Logout button */}
          <button
            onClick={() => supabase.auth.signOut()}
            className="w-9 h-9 rounded-xl hover:bg-surface-800 flex items-center justify-center text-surface-400 hover:text-surface-200 transition-colors"
            title="Sign Out"
          >
            🚪
          </button>
        </div>
      </header>

      {/* Chat area */}
      <ChatWindow messages={messages} isLoading={isLoading} />

      {/* Input */}
      <InputBar onSend={handleSend} disabled={isLoading} />

      {/* Config modal */}
      {showConfig && (
        <ConfigPanel config={config} setConfig={setConfig} session={session} onClose={() => setShowConfig(false)} />
      )}
    </div>
  );
}
