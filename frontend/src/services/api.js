// API service — handles all communication with the Go backend.

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1';

/**
 * Send a chat message to the backend.
 * @param {string} message - The user's message
 * @param {Object} context - Azure DevOps config {organization, project, pat}
 * @returns {Promise<Object>} - Chat response
 */
export async function sendMessage(message, context, token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE}/chat`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ message, context }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({}));
    throw new Error(error.error || `Server error (${response.status})`);
  }

  return response.json();
}

/**
 * Validate Azure DevOps credentials.
 * @param {Object} config - {organization, project, pat}
 * @returns {Promise<Object>} - Validation result
 */
export async function validateConfig(config) {
  const response = await fetch(`${API_BASE}/config/validate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(config),
  });

  return response.json();
}

/**
 * Save user configuration to the cloud database.
 * @param {Object} config - {organization, project, pat}
 * @param {string} token - Session token
 * @returns {Promise<Object>} - Save result
 */
export async function saveUserConfig(config, token) {
  const headers = { 'Content-Type': 'application/json' };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const response = await fetch(`${API_BASE}/config/save`, {
    method: 'POST',
    headers,
    body: JSON.stringify(config),
  });

  return response.json();
}

/**
 * Load user configuration from the cloud database.
 * @param {string} token - Session token
 * @returns {Promise<Object>} - Configuration object
 */
export async function loadUserConfig(token) {
  const headers = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const response = await fetch(`${API_BASE}/config/load`, {
    method: 'GET',
    headers,
  });

  return response.json();
}

/**
 * Check if the backend server is healthy.
 * @returns {Promise<Object>} - Health status
 */
export async function checkHealth() {
  const response = await fetch(`${API_BASE}/health`);
  return response.json();
}
