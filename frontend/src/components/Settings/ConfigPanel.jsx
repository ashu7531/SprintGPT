import { useState } from 'react';
import { validateConfig, saveUserConfig } from '../../services/api';

export default function ConfigPanel({ config, setConfig, session, onClose }) {
  const [formData, setFormData] = useState({
    organization: config.organization || '',
    project: config.project || '',
    pat: config.pat || '',
  });
  const [validating, setValidating] = useState(false);
  const [validationResult, setValidationResult] = useState(null);

  const handleChange = (field, value) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    setValidationResult(null);
  };

  const handleValidate = async () => {
    setValidating(true);
    setValidationResult(null);
    try {
      const result = await validateConfig(formData);
      setValidationResult(result);
    } catch (err) {
      setValidationResult({ valid: false, message: err.message });
    }
    setValidating(false);
  };

  const handleSave = async () => {
    setConfig(formData);
    localStorage.setItem('sprintgpt-config', JSON.stringify(formData));
    if (session?.access_token) {
      try { await saveUserConfig(formData, session.access_token); }
      catch (err) { console.error('Failed to sync config:', err); }
    }
    onClose();
  };

  const isComplete = formData.organization && formData.project && formData.pat;

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        inset: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.7)',
        backdropFilter: 'blur(4px)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 50,
        padding: '16px',
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          backgroundColor: '#1f1f1f',
          border: '1px solid #404040',
          borderRadius: '16px',
          width: '100%',
          maxWidth: '440px',
          overflow: 'hidden',
        }}
      >
        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '20px 24px', borderBottom: '1px solid #2e2e2e' }}>
          <div>
            <h2 style={{ fontSize: '16px', fontWeight: 600, color: '#f5f5f5', margin: 0 }}>Azure DevOps Configuration</h2>
            <p style={{ fontSize: '13px', color: '#737373', margin: '4px 0 0' }}>Connect to your project</p>
          </div>
          <button
            onClick={onClose}
            style={{ width: '32px', height: '32px', borderRadius: '8px', border: 'none', backgroundColor: '#262626', color: '#a3a3a3', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '18px' }}
          >
            ×
          </button>
        </div>

        {/* Form */}
        <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '20px' }}>
          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: '#d4d4d4', marginBottom: '8px' }}>Organization</label>
            <input
              type="text"
              value={formData.organization}
              onChange={(e) => handleChange('organization', e.target.value)}
              placeholder="e.g. my-company"
              style={{ width: '100%', padding: '10px 14px', borderRadius: '10px', border: '1px solid #404040', backgroundColor: '#171717', color: '#f5f5f5', fontSize: '14px', outline: 'none', fontFamily: 'inherit' }}
            />
            <p style={{ fontSize: '11px', color: '#737373', margin: '6px 0 0' }}>From: dev.azure.com/<span style={{ color: '#a3a3a3' }}>your-org</span></p>
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: '#d4d4d4', marginBottom: '8px' }}>Project</label>
            <input
              type="text"
              value={formData.project}
              onChange={(e) => handleChange('project', e.target.value)}
              placeholder="e.g. MyProject"
              style={{ width: '100%', padding: '10px 14px', borderRadius: '10px', border: '1px solid #404040', backgroundColor: '#171717', color: '#f5f5f5', fontSize: '14px', outline: 'none', fontFamily: 'inherit' }}
            />
          </div>

          <div>
            <label style={{ display: 'block', fontSize: '13px', fontWeight: 500, color: '#d4d4d4', marginBottom: '8px' }}>Personal Access Token</label>
            <input
              type="password"
              value={formData.pat}
              onChange={(e) => handleChange('pat', e.target.value)}
              placeholder="Paste your PAT here"
              style={{ width: '100%', padding: '10px 14px', borderRadius: '10px', border: '1px solid #404040', backgroundColor: '#171717', color: '#f5f5f5', fontSize: '14px', outline: 'none', fontFamily: 'inherit' }}
            />
          </div>

          {validationResult && (
            <p style={{
              fontSize: '13px',
              padding: '10px 14px',
              borderRadius: '10px',
              backgroundColor: validationResult.valid ? 'rgba(52, 211, 153, 0.1)' : 'rgba(248, 113, 113, 0.1)',
              border: `1px solid ${validationResult.valid ? 'rgba(52, 211, 153, 0.3)' : 'rgba(248, 113, 113, 0.3)'}`,
              color: validationResult.valid ? '#34d399' : '#f87171',
              margin: 0,
            }}>
              {validationResult.valid ? '✓ ' : '✕ '}{validationResult.message}
            </p>
          )}
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', gap: '12px', padding: '16px 24px 24px', borderTop: '1px solid #2e2e2e' }}>
          <button
            onClick={handleValidate}
            disabled={!isComplete || validating}
            style={{ flex: 1, padding: '12px', borderRadius: '10px', border: '1px solid #404040', backgroundColor: 'transparent', color: '#d4d4d4', fontSize: '13px', fontWeight: 500, cursor: isComplete && !validating ? 'pointer' : 'not-allowed', opacity: isComplete ? 1 : 0.4, fontFamily: 'inherit' }}
          >
            {validating ? 'Testing...' : 'Test Connection'}
          </button>
          <button
            onClick={handleSave}
            disabled={!isComplete}
            style={{ flex: 1, padding: '12px', borderRadius: '10px', border: 'none', backgroundColor: '#10b981', color: 'white', fontSize: '13px', fontWeight: 600, cursor: isComplete ? 'pointer' : 'not-allowed', opacity: isComplete ? 1 : 0.4, fontFamily: 'inherit' }}
          >
            Save & Connect
          </button>
        </div>
      </div>
    </div>
  );
}
