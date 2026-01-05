import type { Notifier } from '@story-engine/shared-ts';

export class ToastNotifier implements Notifier {
  private showToast(message: string, type: 'success' | 'error' | 'info' | 'warning') {
    // Simple toast implementation - can be replaced with a toast library
    const toast = document.createElement('div');
    toast.className = `fixed top-4 right-4 p-4 rounded-[var(--se-radius-md)] z-50 ${
      type === 'success'
        ? 'bg-green-500'
        : type === 'error'
        ? 'bg-red-500'
        : type === 'warning'
        ? 'bg-yellow-500'
        : 'bg-blue-500'
    } text-white`;
    toast.textContent = message;
    document.body.appendChild(toast);

    setTimeout(() => {
      toast.remove();
    }, 3000);
  }

  success(message: string, duration?: number): void {
    this.showToast(message, 'success');
  }

  error(message: string, duration?: number): void {
    this.showToast(message, 'error');
  }

  info(message: string, duration?: number): void {
    this.showToast(message, 'info');
  }

  warning(message: string, duration?: number): void {
    this.showToast(message, 'warning');
  }
}

