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
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [inputValue, setInputValue] = useState('');

  useEffect(() => {
    if (session?.access_token) {
      loadUserConfig(session.access_token)
        .then((savedConfig) => {
          if (savedConfig && savedConfig.organization) {
            setConfig(savedConfig);
          } else {
            const saved = localStorage.getItem('sprintgpt-config');
            if (saved) {
              try { setConfig(JSON.parse(saved)); } catch (e) {}
            }
          }
        })
        .catch((err) => console.error('Failed to load config:', err));
    }
  }, [session]);

  useEffect(() => {
    supabase.auth.getSession().then(({ data: { session } }) => setSession(session));
    const { data: { subscription } } = supabase.auth.onAuthStateChange((_event, session) => setSession(session));
    return () => subscription.unsubscribe();
  }, []);

  useEffect(() => {
    checkHealth()
      .then(() => setServerStatus('online'))
      .catch(() => setServerStatus('offline'));
  }, []);

  const isConfigured = config.organization && config.project && config.pat;

  const handleSend = async (message) => {
    const userMsg = { role: 'user', content: message };
    setMessages((prev) => [...prev, userMsg]);
    setIsLoading(true);
    try {
      const response = await sendMessage(message, config, session?.access_token);
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: response.response, intent: response.intent, data: response.data },
      ]);
    } catch (error) {
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: `**Error:** ${error.message}`, intent: 'error' },
      ]);
    }
    setIsLoading(false);
  };

  if (!session) {
    return <Auth onAuthSuccess={(sess) => setSession(sess)} />;
  }

  return (
    <div style={{ height: '100%', display: 'flex', backgroundColor: '#171717' }}>
      {/* Sidebar */}
      <aside style={{
        width: sidebarOpen ? '260px' : '0px',
        flexShrink: 0,
        backgroundColor: '#0f0f0f',
        borderRight: sidebarOpen ? '1px solid #262626' : 'none',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        transition: 'width 0.2s ease',
      }}>
        {/* Sidebar header */}
        <div style={{ padding: '16px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
            <div style={{ width: '32px', height: '32px', borderRadius: '8px', backgroundColor: '#10b981', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="white" style={{ width: '16px', height: '16px' }}>
                <path d="M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z" />
              </svg>
            </div>
            <span style={{ fontSize: '14px', fontWeight: 600, color: '#e5e5e5' }}>SprintGPT</span>
          </div>
        </div>

        {/* Spacer */}
        <div style={{ flex: 1 }} />

        {/* Bottom section */}
        <div style={{ padding: '12px', borderTop: '1px solid #262626', display: 'flex', flexDirection: 'column', gap: '4px' }}>
          {/* Connection status */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '8px 12px', borderRadius: '8px' }}>
            <div style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: serverStatus === 'online' ? '#34d399' : serverStatus === 'offline' ? '#f87171' : '#fbbf24' }} />
            <span style={{ fontSize: '12px', color: '#a3a3a3' }}>
              {serverStatus === 'online' ? 'Connected' : serverStatus === 'offline' ? 'Offline' : 'Connecting...'}
            </span>
          </div>

          {/* Project info */}
          {isConfigured && (
            <div style={{ padding: '8px 12px', borderRadius: '8px', fontSize: '12px', color: '#737373' }}>
              {config.organization}/{config.project}
            </div>
          )}

          {/* Settings */}
          <button
            id="settings-button"
            onClick={() => setShowConfig(true)}
            style={{ display: 'flex', alignItems: 'center', gap: '10px', width: '100%', padding: '10px 12px', borderRadius: '8px', border: 'none', backgroundColor: 'transparent', color: '#a3a3a3', fontSize: '13px', cursor: 'pointer', textAlign: 'left' }}
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" style={{ width: '16px', height: '16px' }}>
              <path fillRule="evenodd" d="M7.84 1.804A1 1 0 0 1 8.82 1h2.36a1 1 0 0 1 .98.804l.331 1.652a6.993 6.993 0 0 1 1.929 1.115l1.598-.54a1 1 0 0 1 1.186.447l1.18 2.044a1 1 0 0 1-.205 1.251l-1.267 1.113a7.047 7.047 0 0 1 0 2.228l1.267 1.113a1 1 0 0 1 .206 1.25l-1.18 2.045a1 1 0 0 1-1.187.447l-1.598-.54a6.993 6.993 0 0 1-1.929 1.115l-.33 1.652a1 1 0 0 1-.98.804H8.82a1 1 0 0 1-.98-.804l-.331-1.652a6.993 6.993 0 0 1-1.929-1.115l-1.598.54a1 1 0 0 1-1.186-.447l-1.18-2.044a1 1 0 0 1 .205-1.251l1.267-1.114a7.05 7.05 0 0 1 0-2.227L1.821 7.773a1 1 0 0 1-.206-1.25l1.18-2.045a1 1 0 0 1 1.187-.447l1.598.54A6.992 6.992 0 0 1 7.51 3.456l.33-1.652ZM10 13a3 3 0 1 0 0-6 3 3 0 0 0 0 6Z" clipRule="evenodd" />
            </svg>
            Settings
          </button>

          {/* Logout */}
          <button
            onClick={() => supabase.auth.signOut()}
            style={{ display: 'flex', alignItems: 'center', gap: '10px', width: '100%', padding: '10px 12px', borderRadius: '8px', border: 'none', backgroundColor: 'transparent', color: '#a3a3a3', fontSize: '13px', cursor: 'pointer', textAlign: 'left' }}
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" style={{ width: '16px', height: '16px' }}>
              <path fillRule="evenodd" d="M3 4.25A2.25 2.25 0 0 1 5.25 2h5.5A2.25 2.25 0 0 1 13 4.25v2a.75.75 0 0 1-1.5 0v-2a.75.75 0 0 0-.75-.75h-5.5a.75.75 0 0 0-.75.75v11.5c0 .414.336.75.75.75h5.5a.75.75 0 0 0 .75-.75v-2a.75.75 0 0 1 1.5 0v2A2.25 2.25 0 0 1 10.75 18h-5.5A2.25 2.25 0 0 1 3 15.75V4.25Z" clipRule="evenodd" />
              <path fillRule="evenodd" d="M19 10a.75.75 0 0 0-.75-.75H8.704l1.048-.943a.75.75 0 1 0-1.004-1.114l-2.5 2.25a.75.75 0 0 0 0 1.114l2.5 2.25a.75.75 0 1 0 1.004-1.114l-1.048-.943h9.546A.75.75 0 0 0 19 10Z" clipRule="evenodd" />
            </svg>
            Sign Out
          </button>
        </div>
      </aside>

      {/* Main content */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', minWidth: 0 }}>
        {/* Slim top bar with toggle */}
        <div style={{ flexShrink: 0, padding: '12px 24px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <button
            onClick={() => setSidebarOpen(!sidebarOpen)}
            style={{ padding: '6px', borderRadius: '6px', border: 'none', backgroundColor: 'transparent', color: '#737373', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
          >
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" style={{ width: '20px', height: '20px' }}>
              <path fillRule="evenodd" d="M2 4.75A.75.75 0 0 1 2.75 4h14.5a.75.75 0 0 1 0 1.5H2.75A.75.75 0 0 1 2 4.75Zm0 10.5a.75.75 0 0 1 .75-.75h14.5a.75.75 0 0 1 0 1.5H2.75a.75.75 0 0 1-.75-.75ZM2 10a.75.75 0 0 1 .75-.75h7.5a.75.75 0 0 1 0 1.5h-7.5A.75.75 0 0 1 2 10Z" clipRule="evenodd" />
            </svg>
          </button>
          
          {!isConfigured && (
            <div style={{ backgroundColor: 'rgba(251, 191, 36, 0.1)', border: '1px solid rgba(251, 191, 36, 0.3)', color: '#fbbf24', padding: '8px 16px', borderRadius: '8px', fontSize: '13px', display: 'flex', alignItems: 'center', gap: '8px' }}>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" style={{ width: '16px', height: '16px' }}>
                <path fillRule="evenodd" d="M18 10a8 8 0 1 1-16 0 8 8 0 0 1 16 0Zm-7-4a1 1 0 1 1-2 0 1 1 0 0 1 2 0ZM9 9a.75.75 0 0 0 0 1.5h.253a.25.25 0 0 1 .244.304l-.459 2.066A1.75 1.75 0 0 0 10.747 15H11a.75.75 0 0 0 0-1.5h-.253a.25.25 0 0 1-.244-.304l.459-2.066A1.75 1.75 0 0 0 9.253 9H9Z" clipRule="evenodd" />
              </svg>
              <span>By default, a demo project is connected. Please update your Settings to connect your real project.</span>
            </div>
          )}
        </div>

        {/* Chat */}
        <ChatWindow messages={messages} isLoading={isLoading} onSuggestionClick={(text) => setInputValue(text)} />

        {/* Input */}
        <InputBar onSend={handleSend} disabled={isLoading} externalValue={inputValue} onExternalValueUsed={() => setInputValue('')} />
      </div>

      {/* Config modal */}
      {showConfig && (
        <ConfigPanel config={config} setConfig={setConfig} session={session} onClose={() => setShowConfig(false)} />
      )}
    </div>
  );
}
