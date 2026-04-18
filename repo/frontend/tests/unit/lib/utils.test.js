import { cn } from '../../../src/lib/utils';

describe('cn (classname utility)', () => {
  test('returns empty string when called with no args', () => {
    expect(cn()).toBe('');
  });

  test('returns a single class name unchanged', () => {
    expect(cn('text-sm')).toBe('text-sm');
  });

  test('joins multiple class names', () => {
    expect(cn('a', 'b', 'c')).toBe('a b c');
  });

  test('resolves conflicting tailwind classes (last wins)', () => {
    expect(cn('text-sm', 'text-lg')).toBe('text-lg');
    expect(cn('p-2', 'p-4')).toBe('p-4');
  });

  test('supports conditional classes via object syntax', () => {
    expect(cn({ active: true, inactive: false })).toBe('active');
  });

  test('handles falsy values without error', () => {
    expect(cn(undefined, null, false, 'real-class')).toBe('real-class');
  });

  test('handles array input from clsx', () => {
    expect(cn(['a', 'b'])).toBe('a b');
  });

  test('handles conditional with ternary', () => {
    const isActive = true;
    expect(cn('base', isActive ? 'active' : 'idle')).toBe('base active');
  });

  test('merges bg- color conflicts correctly', () => {
    expect(cn('bg-red-500', 'bg-blue-500')).toBe('bg-blue-500');
  });
});
