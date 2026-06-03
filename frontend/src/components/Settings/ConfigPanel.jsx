import { useState } from 'react';
import { validateConfig } from '../../services/api';

export default function ConfigPanel({ config, setConfig, onClose }) {
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

  const handleSave = () => {
    setConfig(formData);
    // Save to localStorage for persistence
    localStorage.setItem('sprintgpt-config', JSON.stringify(formData));
    onClose();
  };

  const isComplete = formData.organization && formData.project && formData.pat;

  return (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-surface-900 border border-surface-700/50 rounded-2xl w-full max-w-md shadow-2xl">
        {/* Header */}
        <div className="flex items-center justify-between p-5 border-b border-surface-800">
          <div>
            <h2 className="text-lg font-semibold text-surface-100">Azure DevOps Configuration</h2>
            <p className="text-sm text-surface-400 mt-0.5">Connect to your project</p>
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-lg hover:bg-surface-800 flex items-center justify-center text-surface-400 hover:text-surface-200 transition-colors"
          >
            ✕
          </button>
        </div>

        {/* Form */}
        <div className="p-5 space-y-4">
          <div>
            <label htmlFor="config-org" className="block text-sm font-medium text-surface-300 mb-1.5">
              Organization
            </label>
            <input
              id="config-org"
              type="text"
              value={formData.organization}
              onChange={(e) => handleChange('organization', e.target.value)}
              placeholder="e.g. my-company"
              className="w-full px-3 py-2.5 bg-surface-800 border border-surface-700 rounded-xl text-sm text-surface-100 placeholder:text-surface-500 outline-none focus:border-primary-500/50 transition-colors"
            />
            <p className="text-xs text-surface-500 mt-1">
              From: dev.azure.com/<strong className="text-surface-400">your-org</strong>
            </p>
          </div>

          <div>
            <label htmlFor="config-project" className="block text-sm font-medium text-surface-300 mb-1.5">
              Project
            </label>
            <input
              id="config-project"
              type="text"
              value={formData.project}
              onChange={(e) => handleChange('project', e.target.value)}
              placeholder="e.g. MyProject"
              className="w-full px-3 py-2.5 bg-surface-800 border border-surface-700 rounded-xl text-sm text-surface-100 placeholder:text-surface-500 outline-none focus:border-primary-500/50 transition-colors"
            />
          </div>

          <div>
            <label htmlFor="config-pat" className="block text-sm font-medium text-surface-300 mb-1.5">
              Personal Access Token (PAT)
            </label>
            <input
              id="config-pat"
              type="password"
              value={formData.pat}
              onChange={(e) => handleChange('pat', e.target.value)}
              placeholder="Paste your PAT here"
              className="w-full px-3 py-2.5 bg-surface-800 border border-surface-700 rounded-xl text-sm text-surface-100 placeholder:text-surface-500 outline-none focus:border-primary-500/50 transition-colors"
            />
            <p className="text-xs text-surface-500 mt-1">
              Create at: User Settings → Personal Access Tokens
            </p>
          </div>

          {/* Validation result */}
          {validationResult && (
            <div
              className={`p-3 rounded-xl text-sm ${
                validationResult.valid
                  ? 'bg-green-500/10 border border-green-500/30 text-green-400'
                  : 'bg-red-500/10 border border-red-500/30 text-red-400'
              }`}
            >
              {validationResult.valid ? '✅ ' : '❌ '}
              {validationResult.message}
            </div>
          )}
        </div>

        {/* Actions */}
        <div className="flex items-center gap-3 p-5 border-t border-surface-800">
          <button
            onClick={handleValidate}
            disabled={!isComplete || validating}
            className="flex-1 px-4 py-2.5 bg-surface-800 hover:bg-surface-700 disabled:opacity-40 rounded-xl text-sm font-medium text-surface-200 transition-colors disabled:cursor-not-allowed"
          >
            {validating ? 'Validating...' : 'Test Connection'}
          </button>
          <button
            onClick={handleSave}
            disabled={!isComplete}
            className="flex-1 px-4 py-2.5 bg-primary-600 hover:bg-primary-500 disabled:opacity-40 rounded-xl text-sm font-medium text-white transition-colors disabled:cursor-not-allowed"
          >
            Save & Connect
          </button>
        </div>
      </div>
    </div>
  );
}
