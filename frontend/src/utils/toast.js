/**
 * Simple toast notification utility
 * Creates toast notifications that appear in the bottom-right corner
 */

let toastId = 0;
const toasts = [];
let toastContainer = null;

/**
 * Create a toast notification
 * @param {string} message - Toast message
 * @param {string} type - Toast type: 'success', 'error', 'info', 'warning'
 * @param {number} duration - Duration in milliseconds (default: 3000)
 */
export function showToast(message, type = 'info', duration = 3000) {
  // Create toast container if it doesn't exist
  if (!toastContainer) {
    toastContainer = document.createElement('div');
    toastContainer.className = 'toast-stack';
    document.body.appendChild(toastContainer);
  }

  const id = toastId++;
  const toast = document.createElement('div');
  toast.className = `toast ${type}`;
  toast.textContent = message;
  toast.setAttribute('role', 'alert');
  toast.setAttribute('aria-live', 'polite');

  toastContainer.appendChild(toast);

  // Auto-remove after duration
  setTimeout(() => {
    toast.style.animation = 'toastExit 0.2s ease';
    setTimeout(() => {
      if (toast.parentNode) {
        toast.parentNode.removeChild(toast);
      }
      const index = toasts.indexOf(id);
      if (index > -1) {
        toasts.splice(index, 1);
      }
    }, 200);
  }, duration);

  toasts.push(id);
  return id;
}

// Add CSS animation for toast exit if not already present
if (!document.getElementById('toast-styles')) {
  const style = document.createElement('style');
  style.id = 'toast-styles';
  style.textContent = `
    @keyframes toastExit {
      from {
        opacity: 1;
        transform: translateY(0);
      }
      to {
        opacity: 0;
        transform: translateY(12px);
      }
    }
  `;
  document.head.appendChild(style);
}

