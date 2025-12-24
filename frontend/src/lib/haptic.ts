/**
 * Haptic feedback utilities for mobile devices
 */

export interface HapticPattern {
  duration: number;
  delay?: number;
}

export class HapticFeedback {
  private static isSupported(): boolean {
    return 'vibrate' in navigator;
  }

  /**
   * Trigger a single vibration
   * @param duration Duration in milliseconds
   */
  static vibrate(duration: number = 50): void {
    if (!this.isSupported()) return;

    navigator.vibrate(duration);
  }

  /**
   * Trigger a vibration pattern
   * @param pattern Array of durations and pauses
   */
  static vibratePattern(pattern: number[]): void {
    if (!this.isSupported()) return;

    navigator.vibrate(pattern);
  }

  /**
   * Light feedback (quick tap)
   */
  static light(): void {
    this.vibrate(50);
  }

  /**
   * Medium feedback (button press)
   */
  static medium(): void {
    this.vibrate(100);
  }

  /**
   * Heavy feedback (important action)
   */
  static heavy(): void {
    this.vibratePattern([100, 50, 100]);
  }

  /**
   * Success feedback
   */
  static success(): void {
    this.vibratePattern([50, 50, 50, 50, 100]);
  }

  /**
   * Error feedback
   */
  static error(): void {
    this.vibratePattern([200, 100, 200]);
  }

  /**
   * Warning feedback
   */
  static warning(): void {
    this.vibratePattern([100, 50, 100, 50, 100]);
  }

  /**
   * Selection feedback (like picker selection)
   */
  static selection(): void {
    this.vibrate(30);
  }

  /**
   * Stop any ongoing vibration
   */
  static stop(): void {
    if (!this.isSupported()) return;

    navigator.vibrate(0);
  }
}

// Convenience functions
export const haptic = {
  light: HapticFeedback.light.bind(HapticFeedback),
  medium: HapticFeedback.medium.bind(HapticFeedback),
  heavy: HapticFeedback.heavy.bind(HapticFeedback),
  success: HapticFeedback.success.bind(HapticFeedback),
  error: HapticFeedback.error.bind(HapticFeedback),
  warning: HapticFeedback.warning.bind(HapticFeedback),
  selection: HapticFeedback.selection.bind(HapticFeedback),
  stop: HapticFeedback.stop.bind(HapticFeedback),
  vibrate: HapticFeedback.vibrate.bind(HapticFeedback),
  vibratePattern: HapticFeedback.vibratePattern.bind(HapticFeedback)
};