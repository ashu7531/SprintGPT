import { useState } from 'react';
import { supabase } from '../../utils/supabase';

export default function Auth({ onAuthSuccess }) {
  const [isRegistering, setIsRegistering] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const [successMsg, setSuccessMsg] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setErrorMsg('');
    setSuccessMsg('');

    try {
      if (isRegistering) {
        // Sign Up
        const { data, error } = await supabase.auth.signUp({
          email,
          password,
        });
        if (error) throw error;
        
        // Supabase sends a confirmation email by default
        setSuccessMsg('Registration successful! Please check your email for a confirmation link.');
      } else {
        // Sign In
        const { data, error } = await supabase.auth.signInWithPassword({
          email,
          password,
        });
        if (error) throw error;
        
        if (data?.session) {
          onAuthSuccess(data.session);
        }
      }
    } catch (err) {
      setErrorMsg(err.message || 'An error occurred during authentication');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-radial from-surface-900 to-surface-950 px-4">
      {/* Background glowing blobs */}
      <div className="absolute top-1/4 left-1/4 w-80 h-80 bg-primary-500/10 rounded-full blur-3xl animate-pulse pointer-events-none" />
      <div className="absolute bottom-1/4 right-1/4 w-80 h-80 bg-primary-600/10 rounded-full blur-3xl animate-pulse pointer-events-none" />

      {/* Auth Card */}
      <div className="w-full max-w-md p-8 rounded-2xl border border-surface-800 bg-surface-900/60 backdrop-blur-xl shadow-2xl flex flex-col items-center">
        {/* Logo Icon */}
        <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-700 flex items-center justify-center mb-6 logo-glow">
          <span className="text-2xl">⚡</span>
        </div>

        <h2 className="text-2xl font-bold text-surface-500 mb-2">
          {isRegistering ? 'Create your Account' : 'Welcome back'}
        </h2>
        <p className="text-sm text-surface-400 mb-8 text-center">
          {isRegistering
            ? 'Start collaborating and summarizing Azure DevOps sprints'
            : 'Sign in to access SprintGPT AI Assistant'}
        </p>

        <form onSubmit={handleSubmit} className="w-full space-y-5">
          {/* Email input */}
          <div>
            <label className="block text-xs font-semibold text-surface-400 mb-2 uppercase tracking-wide">
              Email Address
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full px-4 py-3 rounded-xl border border-surface-800 bg-surface-950/50 text-surface-100 placeholder-surface-650 focus:outline-none focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-all text-sm"
              placeholder="you@example.com"
            />
          </div>

          {/* Password input */}
          <div>
            <label className="block text-xs font-semibold text-surface-400 mb-2 uppercase tracking-wide">
              Password
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full px-4 py-3 rounded-xl border border-surface-800 bg-surface-950/50 text-surface-100 placeholder-surface-650 focus:outline-none focus:border-primary-500 focus:ring-1 focus:ring-primary-500 transition-all text-sm"
              placeholder="••••••••"
            />
          </div>

          {/* Error and Success notifications */}
          {errorMsg && (
            <div className="p-3 rounded-xl border border-red-500/20 bg-red-500/10 text-xs text-red-400 flex items-start gap-2.5">
              <span>⚠️</span>
              <span>{errorMsg}</span>
            </div>
          )}

          {successMsg && (
            <div className="p-3 rounded-xl border border-green-500/20 bg-green-500/10 text-xs text-green-400 flex items-start gap-2.5">
              <span>✅</span>
              <span>{successMsg}</span>
            </div>
          )}

          {/* Submit Button */}
          <button
            type="submit"
            disabled={loading}
            className="w-full py-3 px-4 rounded-xl bg-gradient-to-r from-primary-600 to-primary-500 hover:from-primary-500 hover:to-primary-400 text-surface-100 text-sm font-semibold shadow-lg shadow-primary-500/20 hover:shadow-primary-500/30 transition-all flex items-center justify-center disabled:opacity-55 active:scale-98"
          >
            {loading ? (
              <span className="w-5 h-5 border-2 border-surface-100 border-t-transparent rounded-full animate-spin" />
            ) : isRegistering ? (
              'Sign Up'
            ) : (
              'Sign In'
            )}
          </button>
        </form>

        {/* Toggle link */}
        <div className="mt-8 text-center text-sm text-surface-400">
          {isRegistering ? 'Already have an account? ' : "Don't have an account? "}
          <button
            type="button"
            onClick={() => {
              setIsRegistering(!isRegistering);
              setErrorMsg('');
              setSuccessMsg('');
            }}
            className="text-primary-400 font-semibold hover:text-primary-300 focus:outline-none underline decoration-2 underline-offset-4"
          >
            {isRegistering ? 'Log In' : 'Sign Up'}
          </button>
        </div>
      </div>
    </div>
  );
}
